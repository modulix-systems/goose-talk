package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/schemas"
)

type AuthService struct {
	usersRepo            gateways.UsersRepo
	notificationsServive gateways.NotificationsService
	tgApi                gateways.TelegramBotAPI
	securityProvider     gateways.SecurityProvider
	otpRepo              gateways.OtpRepo
	otpTTL               time.Duration
	authTokenProvider    gateways.AuthTokenProvider
	authTokenTTL         time.Duration
	sessionsRepo         gateways.UserSessionsRepo
	geoIPApi             gateways.GeoIPApi
	twoFactorAuthRepo    gateways.TwoFactorAuthRepo
}

func New(
	usersRepo gateways.UsersRepo,
	notificationsServive gateways.NotificationsService,
	otpRepo gateways.OtpRepo,
	authTokenProvider gateways.AuthTokenProvider,
	otpTTL time.Duration,
	authTokenTTL time.Duration,
	securityProvider gateways.SecurityProvider,
	tgApi gateways.TelegramBotAPI,
	sessionsRepo gateways.UserSessionsRepo,
	geoIPApi gateways.GeoIPApi,
	twoFactorAuthRepo gateways.TwoFactorAuthRepo,
) *AuthService {
	return &AuthService{
		usersRepo:            usersRepo,
		notificationsServive: notificationsServive,
		otpRepo:              otpRepo,
		otpTTL:               otpTTL,
		authTokenProvider:    authTokenProvider,
		authTokenTTL:         authTokenTTL,
		securityProvider:     securityProvider,
		tgApi:                tgApi,
		sessionsRepo:         sessionsRepo,
		geoIPApi:             geoIPApi,
		twoFactorAuthRepo:    twoFactorAuthRepo,
	}
}

// createOTP generates, hashes and saves hashed otp token to database
// returning plain code and insertion error for caller to handle
func (s *AuthService) createOTP(ctx context.Context, email string, userId int) (string, error) {
	if email == "" && userId == 0 {
		panic("createOTP: email or userId must be provided")
	}
	otpCode := s.securityProvider.GenerateOTPCode()
	hashedOtpCode, err := s.securityProvider.HashPassword(otpCode)
	if err != nil {
		return "", err
	}
	return otpCode, s.otpRepo.InsertOrUpdateCode(
		ctx,
		&entity.OTP{Code: hashedOtpCode, UserEmail: email, UserId: userId},
	)
}

func (s *AuthService) createOrUpdateSession(ctx context.Context, user *entity.User, sessionEnt *entity.UserSession) (*entity.UserSession, error) {
	// Try to find matching session by set of params, if it wasn't found - create new one
	// or update otherwise
	session, err := s.sessionsRepo.GetByParamsMatch(ctx, sessionEnt.IP, sessionEnt.DeviceInfo, user.ID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			session, err = s.sessionsRepo.Insert(ctx, sessionEnt)
			if err != nil {
				return nil, err
			}
			return session, nil
		}
		return nil, err
	}
	// activate session back if it was deactivated and update it's token
	return s.sessionsRepo.UpdateById(
		ctx,
		session.ID,
		&schemas.SessionUpdatePayload{
			DeactivatedAt: &time.Time{},
			LastSeenAt:    time.Now(),
			AccessToken:   session.AccessToken,
		},
	)
}

func (s *AuthService) SignUp(
	ctx context.Context,
	dto *schemas.SignUpSchema,
) (string, *entity.User, error) {
	otp, err := s.otpRepo.GetByEmail(ctx, dto.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", nil, ErrOTPInvalidOrExpired
		}
		return "", nil, err
	}
	if otp.IsExpired(s.otpTTL) {
		return "", nil, ErrOTPInvalidOrExpired
	}
	matched, err := s.securityProvider.ComparePasswords(otp.Code, dto.ConfirmationCode)
	if err != nil {
		return "", nil, err
	}
	if !matched {
		return "", nil, ErrOTPInvalidOrExpired
	}
	hashedPassword, err := s.securityProvider.HashPassword(dto.Password)
	if err != nil {
		return "", nil, err
	}
	user, err := s.usersRepo.Insert(
		ctx,
		&entity.User{
			FirstName: dto.FirstName,
			LastName:  dto.LastName,
			Email:     dto.Email,
			Password:  hashedPassword,
		},
	)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			return "", nil, ErrUserAlreadyExists
		}
		return "", nil, err
	}
	authToken, err := s.authTokenProvider.NewToken(s.authTokenTTL, map[string]any{"uid": user.ID})
	if err != nil {
		return "", nil, err
	}
	displayName := dto.Username
	if dto.FirstName != "" {
		displayName = dto.FirstName
		if dto.LastName != "" {
			displayName = displayName + " " + dto.LastName
		}
	}
	s.notificationsServive.SendGreetingEmail(ctx, user.Email, displayName)
	return authToken, user, nil
}

func (s *AuthService) ConfirmEmail(ctx context.Context, email string) error {
	isExists, err := s.usersRepo.CheckExistsWithEmail(ctx, email)
	if err != nil {
		return err
	}
	if isExists {
		return ErrUserAlreadyExists
	}

	otpCode, err := s.createOTP(ctx, email, 0)
	if err != nil {
		return err
	}
	if err = s.notificationsServive.SendSignUpConfirmationEmail(ctx, email, otpCode); err != nil {
		return err
	}
	return nil
}

// tokenType indicates semantic meaning of token value in authInfo
type tokenType = int

const (
	// AuthTokenType indicates that token is authorization token and no 2fa is required
	AuthTokenType tokenType = iota
	// SignInConfTokenType indicates that token value is intended
	// for sign in confirmation in further 2fa verification
	SignInConfTokenType tokenType = iota
)

type signInToken struct {
	Val string
	Typ tokenType
}

type authInfo struct {
	Token   *signInToken
	User    *entity.User
	Session *entity.UserSession
}

func (s *AuthService) SignIn(ctx context.Context, dto *schemas.SignInSchema) (*authInfo, error) {
	user, err := s.usersRepo.GetByLogin(ctx, dto.Login)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if !user.IsActive {
		return nil, ErrDisabledAccount
	}
	matched, err := s.securityProvider.ComparePasswords(user.Password, dto.Password)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, ErrInvalidCredentials
	}
	if user.Is2FAEnabled() {
		otpCode, err := s.createOTP(ctx, "", user.ID)
		if err != nil {
			return nil, err
		}
		contact := user.TwoFactorAuth.Contact
		switch user.TwoFactorAuth.DeliveryMethod {
		case entity.TWO_FA_EMAIL:
			toEmail := user.Email
			if contact != "" {
				toEmail = contact
			}
			if err = s.notificationsServive.Send2FAEmail(ctx, toEmail, otpCode); err != nil {
				return nil, err
			}
		case entity.TWO_FA_TELEGRAM:
			if err = s.tgApi.SendTextMsg(ctx, contact, fmt.Sprintf("Authorization code: %s", otpCode)); err != nil {
				return nil, err
			}
		case entity.TWO_FA_TOTP_APP:
			return &authInfo{
				User:  user,
				Token: &signInToken{Val: otpCode, Typ: SignInConfTokenType},
			}, nil
		default:
			return nil, ErrUnsupported2FAMethod
		}
		return &authInfo{User: user}, nil
	}
	userLocation, err := s.geoIPApi.GetLocationByIP(dto.ClientIP)
	if err != nil {
		return nil, err
	}
	token, err := s.authTokenProvider.NewToken(s.authTokenTTL, map[string]any{"uid": user.ID})
	if err != nil {
		return nil, err
	}
	session, err := s.createOrUpdateSession(ctx, user, &entity.UserSession{
		UserId:      user.ID,
		DeviceInfo:  dto.DeviceInfo,
		IP:          dto.ClientIP,
		Location:    userLocation,
		AccessToken: token,
	})
	if err != nil {
		return nil, err
	}
	return &authInfo{
		User:    user,
		Session: session,
		Token:   &signInToken{Val: token, Typ: AuthTokenType},
	}, nil
}

func (s *AuthService) Verify2FA(ctx context.Context, dto *schemas.Verify2FASchema) (string, error) {
	otp, err := s.otpRepo.GetByEmail(ctx, dto.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", ErrOTPInvalidOrExpired
		}
		return "", err
	}
	if otp.IsExpired(s.otpTTL) {
		return "", ErrOTPInvalidOrExpired
	}
	otpToCompare := dto.Code
	if dto.TwoFATyp == entity.TWO_FA_TOTP_APP {
		otpToCompare = dto.SignInConfToken
	}
	matched, err := s.securityProvider.ComparePasswords(otp.Code, otpToCompare)
	if err != nil {
		return "", err
	}
	if !matched {
		return "", ErrOTPInvalidOrExpired
	}
	user, err := s.usersRepo.GetByLogin(ctx, dto.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			if err = s.otpRepo.DeleteByEmail(ctx, dto.Email); err != nil {
				return "", err
			}
			return "", ErrUserNotFound
		}
		return "", err
	}
	if !user.IsActive {
		return "", ErrDisabledAccount
	}
	if !user.Is2FAEnabled() {
		return "", Err2FANotEnabled
	}
	if dto.TwoFATyp == entity.TWO_FA_TOTP_APP {
		isValid := s.securityProvider.ValidateTOTP(dto.Code, user.TwoFactorAuth.TotpSecret)
		if !isValid {
			return "", ErrOTPInvalidOrExpired
		}
	}
	token, err := s.authTokenProvider.NewToken(s.authTokenTTL, map[string]any{"uid": user.ID})
	if err != nil {
		return "", err
	}
	userLocation, err := s.geoIPApi.GetLocationByIP(dto.ClientIP)
	if err != nil {
		return "", err
	}
	_, err = s.createOrUpdateSession(ctx, user, &entity.UserSession{
		UserId:      user.ID,
		DeviceInfo:  dto.DeviceInfo,
		IP:          dto.ClientIP,
		Location:    userLocation,
		AccessToken: token,
	})
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) Confirm2FA(ctx context.Context, dto *schemas.Confirm2FASchema) (*entity.TwoFactorAuth, error) {
	if dto.Typ == entity.TWO_FA_TOTP_APP {
		isValid := s.securityProvider.ValidateTOTP(dto.ConfirmationCode, dto.TotpSecret)
		if !isValid {
			return nil, ErrOTPInvalidOrExpired
		}
	} else {
		otp, err := s.otpRepo.GetByUserId(ctx, dto.UserId)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, ErrOTPInvalidOrExpired
			}
			return nil, err
		}
		if otp.IsExpired(s.otpTTL) {
			return nil, ErrOTPInvalidOrExpired
		}
		matched, err := s.securityProvider.ComparePasswords(otp.Code, dto.ConfirmationCode)
		if err != nil {
			return nil, err
		}
		if !matched {
			return nil, ErrOTPInvalidOrExpired
		}
	}
	ent := &entity.TwoFactorAuth{
		UserId:         dto.UserId,
		DeliveryMethod: dto.Typ,
		Enabled:        true,
	}
	if dto.Typ == entity.TWO_FA_EMAIL || dto.Typ == entity.TWO_FA_SMS {
		ent.Contact = dto.Contact
	}
	twoFactorAuth, err := s.twoFactorAuthRepo.Insert(ctx, ent)
	if err != nil {
		return nil, err
	}
	return twoFactorAuth, nil
}

func (s *AuthService) tgSendOtpOnMsgAndUpdateContact(userId int, otpCode string, msgCode string) {
	ctx := context.Background()
	startTime := time.Now()
	attemptsLeft := 100
	for attemptsLeft > 0 {
		msg, err := s.tgApi.GetLatestMsg(ctx)
		if err != nil {
			return
		}
		if msg.DateSent.After(startTime) {
			msgParts := strings.Split(msg.Text, " ")
			if len(msgParts) > 1 && msgParts[1] == msgCode {
				err := s.twoFactorAuthRepo.UpdateContactForUser(ctx, userId, msg.ChatId)
				if err != nil {
					println(err)
				}
				err = s.tgApi.SendTextMsg(ctx, msg.ChatId, fmt.Sprintf("Authorization code: %s", otpCode))
				if err != nil {
					println(err)
				}
			}
		}
		time.Sleep(time.Second)
		attemptsLeft--
	}
}

type TwoFAConnectInfo struct {
	Link       string
	TotpSecret string
}

func (s *AuthService) Add2FA(ctx context.Context, dto *schemas.Add2FASchema) (*TwoFAConnectInfo, error) {
	user, err := s.usersRepo.GetByID(ctx, dto.UserId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if user.TwoFactorAuth != nil {
		return nil, Err2FaAlreadyAdded
	}
	switch dto.Typ {
	case entity.TWO_FA_EMAIL:
		emailRecipient := user.Email
		if dto.Contact != "" {
			emailRecipient = dto.Contact
		}
		otpCode, err := s.createOTP(ctx, "", dto.UserId)
		if err != nil {
			return nil, err
		}
		if err = s.notificationsServive.Send2FAEmail(ctx, emailRecipient, otpCode); err != nil {
			return nil, err
		}
		return nil, nil
	case entity.TWO_FA_TELEGRAM:
		otpCode, err := s.createOTP(ctx, "", dto.UserId)
		if err != nil {
			return nil, err
		}
		tgMsgCode := s.securityProvider.GenerateOTPCode()
		link := s.tgApi.GetStartLinkWithCode(tgMsgCode)
		go s.tgSendOtpOnMsgAndUpdateContact(user.ID, otpCode, tgMsgCode)
		return &TwoFAConnectInfo{Link: link}, nil
	case entity.TWO_FA_TOTP_APP:
		link, secret := s.securityProvider.GenerateTOTPEnrollUrlWithSecret(user.Email)
		return &TwoFAConnectInfo{Link: link, TotpSecret: secret}, nil
	default:
		return nil, ErrUnsupported2FAMethod
	}
}

func (s *AuthService) DeactivateAccount(ctx context.Context, userId int) error {
	user, err := s.usersRepo.UpdateIsActiveById(ctx, userId, false)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	if err := s.notificationsServive.SendAccDeactivationEmail(ctx, user.Email); err != nil {
		return err
	}
	return nil
}

func (s *AuthService) GetActiveSessions(
	ctx context.Context,
	authToken string,
) ([]entity.UserSession, error) {
	var sessions []entity.UserSession
	tokenPayload, err := s.authTokenProvider.ParseClaimsFromToken(authToken)
	if err != nil {
		if errors.Is(err, gateways.ErrExpiredToken) {
			return sessions, ErrExpiredAuthToken
		}
		return sessions, ErrInvalidAuthToken
	}
	sessions, err = s.sessionsRepo.GetAllForUser(ctx, tokenPayload["uid"].(int), true)
	if err != nil {
		return sessions, err
	}
	return sessions, nil
}

func (s *AuthService) DeactivateSession(ctx context.Context, userId int, sessionId int) error {
	if err := s.sessionsRepo.UpdateForUserById(ctx, userId, sessionId, time.Now()); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrSessionNotFound
		}
	}
	return nil
}

func (s *AuthService) PingSession(
	ctx context.Context,
	authToken string,
) (*entity.UserSession, error) {
	session, err := s.sessionsRepo.GetByToken(ctx, authToken)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	if !session.IsActive() {
		return nil, ErrSessionNotFound
	}
	now := time.Now()
	updatePayload := &schemas.SessionUpdatePayload{LastSeenAt: now}
	// if token expired - deactivate session
	if _, err = s.authTokenProvider.ParseClaimsFromToken(authToken); err != nil {
		if !errors.Is(err, gateways.ErrExpiredToken) {
			return nil, err
		}
		updatePayload = &schemas.SessionUpdatePayload{DeactivatedAt: &now}
	}
	session, err = s.sessionsRepo.UpdateById(ctx, session.ID, updatePayload)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	return session, nil
}

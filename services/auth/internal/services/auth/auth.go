package auth

import (
	"context"
	"errors"
	"fmt"
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
	}
}

func (s *AuthService) createOTP(ctx context.Context, forEmail string) (string, error) {
	otpCode := s.securityProvider.GenerateOTPCode()
	hashedOtpCode, err := s.securityProvider.HashPassword(otpCode)
	if err != nil {
		return "", err
	}
	return otpCode, s.otpRepo.InsertOrUpdateCode(ctx, &entity.OTP{Code: hashedOtpCode, UserEmail: forEmail})
}

func (s *AuthService) SignUp(
	ctx context.Context,
	dto *schemas.SignUpSchema,
) (string, *entity.User, error) {
	otp, err := s.otpRepo.GetByEmail(ctx, dto.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", nil, ErrInvalidOtp
		}
		return "", nil, err
	}
	if time.Now().After(otp.UpdatedAt.Add(s.otpTTL)) {
		return "", nil, ErrOtpExpired
	}
	matched, err := s.securityProvider.ComparePasswords(otp.Code, dto.ConfirmationCode)
	if err != nil {
		return "", nil, err
	}
	if !matched {
		return "", nil, ErrInvalidOtp
	}
	hashedPassword, err := s.securityProvider.HashPassword(dto.Password)
	if err != nil {
		return "", nil, err
	}
	user, err := s.usersRepo.Insert(
		ctx, &entity.User{FirstName: dto.FirstName, LastName: dto.LastName, Email: dto.Email, Password: hashedPassword},
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

	otpCode, err := s.createOTP(ctx, email)
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
		otpCode, err := s.createOTP(ctx, user.Email)
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
			return &authInfo{User: user, Token: &signInToken{Val: otpCode, Typ: SignInConfTokenType}}, nil
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
	session, err := s.sessionsRepo.Insert(ctx, &entity.UserSession{
		UserId:      user.ID,
		DeviceInfo:  dto.DeviceInfo,
		IP:          dto.ClientIP,
		Location:    userLocation,
		AccessToken: token,
	})
	return &authInfo{User: user, Session: session, Token: &signInToken{Val: token, Typ: AuthTokenType}}, nil
}

func (s *AuthService) Verify2FA(ctx context.Context, dto *schemas.Verify2FASchema) (string, error) {
	otp, err := s.otpRepo.GetByEmail(ctx, dto.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", ErrInvalidOtp
		}
		return "", err
	}
	if time.Now().After(otp.UpdatedAt.Add(s.otpTTL)) {
		return "", ErrOtpExpired
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
		return "", ErrInvalidOtp
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
			return "", ErrInvalidOrExpiredTOTP
		}
	}
	token, err := s.authTokenProvider.NewToken(s.authTokenTTL, map[string]any{"uid": user.ID})
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *AuthService) DeactivateAccount(ctx context.Context, userId string) error {
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

func (s *AuthService) GetCurrentSession(ctx context.Context, authToken string) (*entity.UserSession, error) {
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
	return session, nil
}

// func (s *AuthService) GetAllActiveSessions(ctx context.Context, userId string) ([]entity.UserSession, error)
// func (s *AuthService) DeactivateSession(ctx context.Context, sessionId string) error

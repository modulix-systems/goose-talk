package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/schemas"
	"github.com/modulix-systems/goose-talk/pkg/logger"
)

type AuthService struct {
	usersRepo            gateways.UsersRepo
	notificationsServive gateways.NotificationsService
	tgApi                gateways.TelegramBotAPI
	securityProvider     gateways.SecurityProvider
	otpRepo              gateways.OtpRepo
	otpTTL               time.Duration
	defaultSessionTTL    time.Duration
	longLivedSessionTTL  time.Duration
	loginTokenTTL        time.Duration
	sessionsRepo         gateways.UserSessionsRepo
	geoIPApi             gateways.GeoIPApi
	twoFactorAuthRepo    gateways.TwoFactorAuthRepo
	loginTokenRepo       gateways.LoginTokenRepo
	loginTokenLen        int
	webAuthnProvider     gateways.WebAuthnProvider
	keyValueStorage      gateways.KeyValueStorage
	log                  logger.Interface
}

func New(
	usersRepo gateways.UsersRepo,
	notificationsServive gateways.NotificationsService,
	otpRepo gateways.OtpRepo,
	otpTTL time.Duration,
	loginTokenTTL time.Duration,
	defaultSessionTTL time.Duration,
	longLivedSessionTTL time.Duration,
	securityProvider gateways.SecurityProvider,
	tgApi gateways.TelegramBotAPI,
	sessionsRepo gateways.UserSessionsRepo,
	geoIPApi gateways.GeoIPApi,
	twoFactorAuthRepo gateways.TwoFactorAuthRepo,
	loginTokenRepo gateways.LoginTokenRepo,
	webAuthnProvider gateways.WebAuthnProvider,
	keyValueStorage gateways.KeyValueStorage,
	log logger.Interface,
) *AuthService {
	return &AuthService{
		usersRepo:            usersRepo,
		notificationsServive: notificationsServive,
		otpRepo:              otpRepo,
		otpTTL:               otpTTL,
		defaultSessionTTL:    defaultSessionTTL,
		longLivedSessionTTL:  longLivedSessionTTL,
		loginTokenTTL:        loginTokenTTL,
		securityProvider:     securityProvider,
		tgApi:                tgApi,
		sessionsRepo:         sessionsRepo,
		geoIPApi:             geoIPApi,
		twoFactorAuthRepo:    twoFactorAuthRepo,
		loginTokenRepo:       loginTokenRepo,
		loginTokenLen:        16,
		webAuthnProvider:     webAuthnProvider,
		keyValueStorage:      keyValueStorage,
		log:                  log,
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

func (s *AuthService) useOTP(ctx context.Context, otp *entity.OTP, compareWith string) error {
	if otp.IsExpired(s.otpTTL) {
		return ErrOTPInvalidOrExpired
	}
	matched, err := s.securityProvider.ComparePasswords(otp.Code, compareWith)
	if err != nil {
		return err
	}
	if !matched {
		return ErrOTPInvalidOrExpired
	}
	if err = s.otpRepo.DeleteByEmailOrUserId(ctx, otp.UserEmail, otp.UserId); err != nil {
		return err
	}
	return nil
}

// newAuthSession inserts a new session or updates existing one based on set of params
// if new session was inserted - sends 'warning' email
func (s *AuthService) newAuthSession(ctx context.Context, user *entity.User, sessionEnt *entity.UserSession) (*entity.UserSession, error) {
	// Try to find matching session by set of params, if it wasn't found - create new one
	// or update otherwise
	session, err := s.sessionsRepo.GetByParamsMatch(ctx, sessionEnt.ClientIdentity.IPAddr, sessionEnt.ClientIdentity.DeviceInfo, user.ID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			session, err = s.sessionsRepo.Insert(ctx, sessionEnt)
			if err != nil {
				return nil, err
			}
			s.notificationsServive.SendSignInNewDeviceEmail(ctx, user.Email, session)
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
		},
	)
}

func (s *AuthService) SignUp(
	ctx context.Context,
	dto *schemas.SignUpSchema,
) (*entity.UserSession, error) {
	otp, err := s.otpRepo.GetByEmail(ctx, dto.Email)
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
	hashedPassword, err := s.securityProvider.HashPassword(dto.Password)
	if err != nil {
		return nil, err
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
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}
	userLocation, err := s.geoIPApi.GetLocationByIP(dto.IPAddr)
	if err != nil {
		return nil, err
	}
	sessionTTL := s.defaultSessionTTL
	if dto.RememberMe {
		sessionTTL = s.longLivedSessionTTL
	}
	authSession, err := s.sessionsRepo.Insert(ctx, &entity.UserSession{
		ID:     s.securityProvider.GenerateSessionId(),
		UserId: user.ID,
		ClientIdentity: &entity.ClientIdentity{
			DeviceInfo: dto.DeviceInfo,
			IPAddr:     dto.IPAddr,
			Location:   userLocation,
		},
		ExpiresAt: time.Now().Add(sessionTTL),
	})
	authSession.User = user
	displayName := dto.Username
	if dto.FirstName != "" {
		displayName = dto.FirstName
		if dto.LastName != "" {
			displayName = displayName + " " + dto.LastName
		}
	}
	s.notificationsServive.SendGreetingEmail(ctx, user.Email, displayName)
	return authSession, nil
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

type authInfo struct {
	SignInConfTokenType string
	User                *entity.User
	Session             *entity.UserSession
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
				User:                user,
				SignInConfTokenType: otpCode,
			}, nil
		default:
			return nil, ErrUnsupported2FAMethod
		}
		return &authInfo{User: user}, nil
	}
	userLocation, err := s.geoIPApi.GetLocationByIP(dto.IPAddr)
	if err != nil {
		return nil, err
	}
	session, err := s.newAuthSession(ctx, user, &entity.UserSession{
		ID:     s.securityProvider.GenerateSessionId(),
		UserId: user.ID,
		ClientIdentity: &entity.ClientIdentity{
			DeviceInfo: dto.DeviceInfo,
			IPAddr:     dto.IPAddr,
			Location:   userLocation,
		},
	})
	if err != nil {
		return nil, err
	}
	return &authInfo{
		User:    user,
		Session: session,
	}, nil
}

func (s *AuthService) Verify2FA(ctx context.Context, dto *schemas.Verify2FASchema) (*entity.UserSession, error) {
	otp, err := s.otpRepo.GetByEmail(ctx, dto.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrOTPInvalidOrExpired
		}
		return nil, err
	}
	if otp.IsExpired(s.otpTTL) {
		return nil, ErrOTPInvalidOrExpired
	}
	otpToCompare := dto.Code
	if dto.TwoFATyp == entity.TWO_FA_TOTP_APP {
		otpToCompare = dto.SignInConfToken
	}
	matched, err := s.securityProvider.ComparePasswords(otp.Code, otpToCompare)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, ErrOTPInvalidOrExpired
	}
	user, err := s.usersRepo.GetByLogin(ctx, otp.UserEmail)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			s.log.Error("Failed to get user by email in existing OTP token", "email", otp.UserEmail)
		}
		return nil, err
	}
	if !user.IsActive {
		return nil, ErrDisabledAccount
	}
	if !user.Is2FAEnabled() {
		return nil, Err2FANotEnabled
	}
	if dto.TwoFATyp == entity.TWO_FA_TOTP_APP {
		decryptedSecret, err := s.securityProvider.DecryptSymmetric(user.TwoFactorAuth.TotpSecret)
		if err != nil {
			return nil, err
		}
		isValid := s.securityProvider.ValidateTOTP(dto.Code, decryptedSecret)
		if !isValid {
			return nil, ErrOTPInvalidOrExpired
		}
	}
	userLocation, err := s.geoIPApi.GetLocationByIP(dto.IPAddr)
	if err != nil {
		return nil, err
	}
	authSession, err := s.newAuthSession(ctx, user, &entity.UserSession{
		ID:     s.securityProvider.GenerateSessionId(),
		UserId: user.ID,
		ClientIdentity: &entity.ClientIdentity{
			DeviceInfo: dto.DeviceInfo,
			IPAddr:     dto.IPAddr,
			Location:   userLocation,
		},
	})
	if err != nil {
		return nil, err
	}
	return authSession, nil
}

func (s *AuthService) Confirm2FaAddition(ctx context.Context, dto *schemas.Confirm2FASchema) (*entity.TwoFactorAuth, error) {
	ent := &entity.TwoFactorAuth{
		UserId:         dto.UserId,
		DeliveryMethod: dto.Typ,
		Enabled:        true,
	}
	if dto.Typ == entity.TWO_FA_TOTP_APP {
		isValid := s.securityProvider.ValidateTOTP(dto.ConfirmationCode, dto.TotpSecret)
		if !isValid {
			return nil, ErrOTPInvalidOrExpired
		}
		encryptedSecret, err := s.securityProvider.EncryptSymmetric(dto.TotpSecret)
		if err != nil {
			return nil, err
		}
		ent.TotpSecret = encryptedSecret
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
	sessionId string,
) ([]entity.UserSession, error) {
	authSession, err := s.sessionsRepo.GetById(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	sessions, err := s.sessionsRepo.GetAllForUser(ctx, authSession.UserId, true)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (s *AuthService) DeactivateSession(ctx context.Context, userId int, sessionId string) error {
	if err := s.sessionsRepo.UpdateForUserById(ctx, userId, sessionId, time.Now()); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrSessionNotFound
		}
	}
	return nil
}

func (s *AuthService) PingSession(
	ctx context.Context,
	sessionId string,
) (*entity.UserSession, error) {
	session, err := s.sessionsRepo.GetById(ctx, sessionId)
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
	// TODO: update expires too
	updatePayload := &schemas.SessionUpdatePayload{LastSeenAt: now}
	session, err = s.sessionsRepo.UpdateById(ctx, session.ID, updatePayload)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *AuthService) ExportLoginToken(ctx context.Context, dto *schemas.ExportLoginTokenSchema) (*entity.LoginToken, error) {
	token, err := s.loginTokenRepo.GetByClientId(ctx, dto.ClientId)
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return nil, err
		}
		tokenValue := s.securityProvider.GenerateSecretTokenUrlSafe(s.loginTokenLen)
		userLocation, err := s.geoIPApi.GetLocationByIP(dto.IPAddr)
		if err != nil {
			return nil, err
		}
		return s.loginTokenRepo.Insert(ctx, &entity.LoginToken{
			ClientId: dto.ClientId,
			Val:      tokenValue,
			ClientIdentity: &entity.ClientIdentity{
				IPAddr:     dto.IPAddr,
				DeviceInfo: dto.DeviceInfo,
				Location:   userLocation,
			},
			ExpiresAt: time.Now().Add(s.loginTokenTTL),
		})
	}
	if token.IsApproved() {
		return token, nil
	}
	err = s.loginTokenRepo.DeleteByClientId(ctx, token.ClientId)
	if err != nil {
		return nil, err
	}
	tokenValue := s.securityProvider.GenerateSecretTokenUrlSafe(s.loginTokenLen)
	return s.loginTokenRepo.Insert(ctx, &entity.LoginToken{
		ClientId:         dto.ClientId,
		Val:              tokenValue,
		ClientIdentityId: token.ClientIdentityId, // reuse previous identity from existent token
		ExpiresAt:        time.Now().Add(s.loginTokenTTL),
	})
}

// AcceptLoginToken is used to authenticated another device from an authorized one (qrcode auth)
func (s *AuthService) AcceptLoginToken(ctx context.Context, userId int, tokenVal string) (*entity.UserSession, error) {
	token, err := s.loginTokenRepo.GetByValue(ctx, tokenVal)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrInvalidLoginToken
		}
		return nil, err
	}
	if token.IsExpired() {
		return nil, ErrExpiredLoginToken
	}
	user, err := s.usersRepo.GetByID(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	session, err := s.newAuthSession(ctx, user, &entity.UserSession{
		ID:               s.securityProvider.GenerateSessionId(),
		UserId:           user.ID,
		ClientIdentityId: token.ClientIdentityId,
		ClientIdentity:   token.ClientIdentity,
	})
	if err != nil {
		return nil, err
	}
	if err := s.loginTokenRepo.UpdateAuthSessionByClientId(ctx, token.ClientId, session.ID); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *AuthService) getPasskeySessionKey(userId int) string {
	return fmt.Sprintf("passkey_session:%d", userId)
}

func (s *AuthService) BeginPasskeyRegistration(ctx context.Context, userId int) (gateways.WebAuthnRegistrationOptions, error) {
	user, err := s.usersRepo.GetByID(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	registrationOptions, passkeySession, err := s.webAuthnProvider.GenerateRegistrationOptions(user)
	if err != nil {
		return nil, err
	}
	serializedPasskeySession, err := json.Marshal(passkeySession)
	if err != nil {
		return nil, err
	}
	if err := s.keyValueStorage.Set(s.getPasskeySessionKey(user.ID), string(serializedPasskeySession), 0); err != nil {
		return nil, err
	}
	return registrationOptions, nil
}

func (s *AuthService) FinishPasskeyRegistration(ctx context.Context, userId int, rawCredential []byte) error {
	passkeySessionJson, err := s.keyValueStorage.Get(s.getPasskeySessionKey(userId))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrPasskeyRegistrationNotInProgress
		}
		return err
	}
	var passkeySession gateways.PasskeyTmpSession
	if err := json.Unmarshal([]byte(passkeySessionJson), &passkeySession); err != nil {
		return err
	}
	cred, err := s.webAuthnProvider.VerifyRegistrationOptions(userId, rawCredential, &passkeySession)
	if err != nil {
		if errors.Is(err, gateways.ErrInvalidCredential) {
			return ErrInvalidPasskeyCredential
		}
		return err
	}
	if err := s.usersRepo.AddPasskeyCredential(ctx, userId, cred); err != nil {
		return err
	}
	return nil
}

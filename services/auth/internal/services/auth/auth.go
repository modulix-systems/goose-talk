package auth

import (
	"context"
	"errors"
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
	SessionTTLThreshold  time.Duration
	SessionTTLAddend     time.Duration
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
		SessionTTLThreshold:  30 * time.Minute,
		SessionTTLAddend:     15 * time.Minute,
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

func (s *AuthService) checkOTP(ctx context.Context, otp *entity.OTP, compareWith string) error {
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
	return nil
}

func (s *AuthService) removeOTP(ctx context.Context, otp *entity.OTP) error {
	return s.otpRepo.DeleteByEmailOrUserId(ctx, otp.UserEmail, otp.UserId)
}

// newAuthSession inserts a new session or updates existing one based on set of params
// if new session was inserted - sends 'warning' email
func (s *AuthService) newAuthSession(ctx context.Context, user *entity.User, sessionEnt *entity.UserSession, rememberMe bool) (*entity.UserSession, error) {
	// Try to find matching session by set of params, if it wasn't found - create new one
	// or update otherwise
	session, err := s.sessionsRepo.GetByParamsMatch(ctx, sessionEnt.ClientIdentity.IPAddr, sessionEnt.ClientIdentity.DeviceInfo, user.ID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			session, err = s.sessionsRepo.Insert(ctx, sessionEnt)
			if err != nil {
				return nil, err
			}
			if err = s.notificationsServive.SendSignInNewDeviceEmail(ctx, user.Email, session); err != nil {
				s.log.Error("Failed to send 'sign in from new device' notification after creating new session", "sessionId", session.ID)
				return nil, err
			}
			return session, nil
		}
		return nil, err
	}
	// activate session back if it was deactivated and update it's token
	sessionTTL := s.defaultSessionTTL
	if rememberMe {
		sessionTTL = s.longLivedSessionTTL
	}
	return s.sessionsRepo.UpdateById(
		ctx,
		session.ID,
		&schemas.SessionUpdatePayload{
			DeactivatedAt: &time.Time{},
			LastSeenAt:    time.Now(),
			ExpiresAt:     time.Now().Add(sessionTTL),
		},
	)
}

package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/pkg/logger"
)

type AuthService struct {
	usersRepo           gateways.UsersRepo
	notificationsClient gateways.NotificationsClient
	tgApi               gateways.TelegramBotClient
	securityProvider    gateways.SecurityProvider
	otpRepo             gateways.OtpRepo
	passkeySessionsRepo gateways.PasskeySessionsRepo
	otpTTL              time.Duration
	defaultSessionTTL   time.Duration
	longLivedSessionTTL time.Duration
	loginTokenTTL       time.Duration
	sessionsRepo        gateways.AuthSessionsRepo
	geoIPApi            gateways.GeoIPApi
	loginTokenRepo      gateways.QRLoginTokenRepo
	webAuthnProvider    gateways.WebAuthnProvider
	log                 logger.Interface
}

func New(
	usersRepo gateways.UsersRepo,
	sessionsRepo gateways.AuthSessionsRepo,
	loginTokenRepo gateways.QRLoginTokenRepo,
	otpRepo gateways.OtpRepo,
	passkeySessionRepo gateways.PasskeySessionsRepo,

	notificationsClient gateways.NotificationsClient,
	webAuthnProvider gateways.WebAuthnProvider,
	securityProvider gateways.SecurityProvider,
	tgApi gateways.TelegramBotClient,
	geoIPApi gateways.GeoIPApi,

	otpTTL time.Duration,
	loginTokenTTL time.Duration,
	defaultSessionTTL time.Duration,
	longLivedSessionTTL time.Duration,

	log logger.Interface,
) *AuthService {
	return &AuthService{
		usersRepo:           usersRepo,
		passkeySessionsRepo: passkeySessionRepo,
		notificationsClient: notificationsClient,
		otpRepo:             otpRepo,
		otpTTL:              otpTTL,
		defaultSessionTTL:   defaultSessionTTL,
		longLivedSessionTTL: longLivedSessionTTL,
		loginTokenTTL:       loginTokenTTL,
		securityProvider:    securityProvider,
		tgApi:               tgApi,
		sessionsRepo:        sessionsRepo,
		geoIPApi:            geoIPApi,
		loginTokenRepo:      loginTokenRepo,
		webAuthnProvider:    webAuthnProvider,
		log:                 log,
	}
}

// createOtp generates, hashes and saves hashed otp token to database
// returning plain code and insertion error for caller to handle
func (s *AuthService) createOtp(ctx context.Context, email string, userId int) (string, error) {
	if email == "" && userId == 0 {
		panic("AuthService - createOtp - email or userId must be provided")
	}

	plainCode := s.securityProvider.GenerateOTPCode()

	hashedCode, err := s.securityProvider.HashPassword(plainCode)
	if err != nil {
		return "", err
	}

	otp := &entity.OTP{Code: hashedCode, UserEmail: email, UserId: userId}

	return plainCode, s.otpRepo.CreateWithTTL(ctx, otp, s.otpTTL)
}

// newAuthSession inserts a new session or updates existing one based on set of params
// if new session was created - sends 'warning' email
func (s *AuthService) newAuthSession(ctx context.Context, user *entity.User, ip string, deviceInfo string, rememberMe bool) (*entity.AuthSession, error) {
	existingSession, err := s.sessionsRepo.GetByLoginData(ctx, user.ID, ip, deviceInfo)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, err
	}

	if existingSession != nil {
		if err := s.sessionsRepo.DeleteById(ctx, user.ID, existingSession.ID); err != nil {
			return nil, err
		}
	}

	sessionTTL := s.defaultSessionTTL
	if rememberMe {
		sessionTTL = s.longLivedSessionTTL
	}

	location, err := s.geoIPApi.GetLocationByIP(ip)
	if err != nil {
		return nil, err
	}

	newSession, err := s.sessionsRepo.CreateWithTTL(
		ctx,
		&entity.AuthSession{
			ID:          s.securityProvider.GenerateSessionId(),
			UserId:      user.ID,
			IPAddr:      ip,
			DeviceInfo:  deviceInfo,
			Location:    location,
			IsLongLived: rememberMe,
		},
		sessionTTL,
	)
	if err != nil {
		return nil, err
	}

	if existingSession == nil {
		if err = s.notificationsClient.SendSignInNewDeviceEmail(ctx, user.Email, newSession); err != nil {
			s.log.Error(fmt.Errorf("AuthService - newAuthSession - notificationsClient.SendSignInNewDeviceEmail: %w", err), "sessionID", newSession.ID)
			return nil, err
		}
	}

	return newSession, nil
}

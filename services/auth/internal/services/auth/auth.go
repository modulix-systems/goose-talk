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
	}
}

func (s *AuthService) createOTP(ctx context.Context, forEmail string) (string, error) {
	otpCode := s.securityProvider.GenerateOTPCode(6)
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
	hashedOtp, err := s.otpRepo.GetByEmail(ctx, dto.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", nil, ErrInvalidSignUpCode
		}
		return "", nil, err
	}
	if time.Now().After(hashedOtp.UpdatedAt.Add(s.otpTTL)) {
		return "", nil, ErrExpiredSignUpCode
	}
	matched, err := s.securityProvider.ComparePasswords(hashedOtp.Code, dto.ConfirmationCode)
	if err != nil {
		return "", nil, err
	}
	if !matched {
		return "", nil, ErrInvalidSignUpCode
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

func (s *AuthService) SignIn(ctx context.Context, dto *schemas.SignInSchema) (string, *entity.User, error) {
	user, err := s.usersRepo.GetByLogin(ctx, dto.Login)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}
	if !user.IsActive {
		return "", nil, ErrDisabledAccount
	}
	matched, err := s.securityProvider.ComparePasswords(user.Password, dto.Password)
	if err != nil {
		return "", nil, err
	}
	if !matched {
		return "", nil, ErrInvalidCredentials
	}
	if user.TwoFactorAuth != nil && user.TwoFactorAuth.Enabled {
		otpCode, err := s.createOTP(ctx, user.Email)
		if err != nil {
			return "", nil, err
		}
		contact := user.TwoFactorAuth.Contact
		switch user.TwoFactorAuth.DeliveryMethod {
		case entity.TWO_FA_EMAIL:
			toEmail := user.Email
			if contact != "" {
				toEmail = contact
			}
			if err = s.notificationsServive.Send2FAEmail(ctx, toEmail, otpCode); err != nil {
				return "", nil, err
			}
		case entity.TWO_FA_TELEGRAM:
			if err = s.tgApi.SendTextMsg(ctx, contact, fmt.Sprintf("Authorization code: %s", otpCode)); err != nil {
				return "", nil, err
			}
		default:
			return "", nil, ErrUnsupported2FAMethod
		}
		return "", user, nil
	}
	authToken, err := s.authTokenProvider.NewToken(s.authTokenTTL, map[string]any{"uid": user.ID})
	if err != nil {
		return "", nil, err
	}
	return authToken, user, nil

}

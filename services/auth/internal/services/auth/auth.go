package auth

import (
	"context"
	"errors"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/schemas"
)

type AuthService struct {
	usersRepo            gateways.UsersRepo
	notificationsServive gateways.NotificationsService
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
) *AuthService {
	return &AuthService{
		usersRepo:            usersRepo,
		notificationsServive: notificationsServive,
		otpRepo:              otpRepo,
		otpTTL:               otpTTL,
		authTokenProvider:    authTokenProvider,
		authTokenTTL:         authTokenTTL,
		securityProvider:     securityProvider,
	}
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
	hashedOtp := s.securityProvider.GenerateOTPCode(6)
	hashedCode, err := s.securityProvider.HashPassword(hashedOtp)
	if err = s.otpRepo.InsertOrUpdateCode(ctx, &entity.OTP{Code: hashedCode, UserEmail: email}); err != nil {
		return err
	}
	if err = s.notificationsServive.SendSignUpConfirmationEmail(ctx, email, hashedOtp); err != nil {
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
	matched, err := s.securityProvider.ComparePasswords(user.Password, dto.Password)
	if err != nil {
		return "", nil, err
	}
	if !matched {
		return "", nil, ErrInvalidCredentials
	}
	authToken, err := s.authTokenProvider.NewToken(s.authTokenTTL, map[string]any{"uid": user.ID})
	if err != nil {
		return "", nil, err
	}
	return authToken, user, nil

}

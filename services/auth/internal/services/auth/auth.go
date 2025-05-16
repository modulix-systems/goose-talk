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
	signUpCodeRepo       gateways.SignUpCodeRepo
	signUpCodeTTL        time.Duration
	authTokenProvider    gateways.AuthTokenProvider
	authTokenTTL         time.Duration
}

func New(
	usersRepo gateways.UsersRepo,
	notificationsServive gateways.NotificationsService,
	signUpCodeRepo gateways.SignUpCodeRepo,
	authTokenProvider gateways.AuthTokenProvider,
	signUpCodeTTL time.Duration,
	authTokenTTL time.Duration,
	securityProvider gateways.SecurityProvider,
) *AuthService {
	return &AuthService{
		usersRepo:            usersRepo,
		notificationsServive: notificationsServive,
		signUpCodeRepo:       signUpCodeRepo,
		signUpCodeTTL:        signUpCodeTTL,
		authTokenProvider:    authTokenProvider,
		authTokenTTL:         authTokenTTL,
		securityProvider:     securityProvider,
	}
}

func (s *AuthService) SignUp(
	ctx context.Context,
	dto *schemas.SignUpSchema,
) (string, *entity.User, error) {
	code, err := s.signUpCodeRepo.GetByEmail(ctx, dto.Email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", nil, ErrInvalidSignUpCode
		}
		return "", nil, err
	}
	if time.Now().After(code.CreatedAt.Add(s.signUpCodeTTL)) {
		return "", nil, ErrExpiredSignUpCode
	}
	user, err := s.usersRepo.Insert(
		ctx, &entity.User{FirstName: dto.FirstName, LastName: dto.LastName, Email: dto.Email},
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
	code := s.securityProvider.NewSecureToken(6)
	if err = s.signUpCodeRepo.Insert(ctx, &entity.SignUpCode{Code: code, Email: email}); err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			signUpCode, err := s.signUpCodeRepo.GetByEmail(ctx, email)
			if err != nil {
				return err
			}
			code = signUpCode.Code
		} else {
			return err
		}
	}
	if err = s.notificationsServive.SendSignUpConfirmationEmail(ctx, email, code); err != nil {
		return err
	}
	return nil
}

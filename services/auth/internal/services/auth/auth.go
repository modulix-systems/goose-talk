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
	usersRepo         gateways.UsersRepo
	signUpCodeRepo    gateways.SignUpCodeRepo
	signUpCodeTTL     time.Duration
	authTokenProvider gateways.AuthTokenProvider
	authTokenTTL      time.Duration
}

func New(
	usersRepo gateways.UsersRepo,
	signUpCodeRepo gateways.SignUpCodeRepo,
	authTokenProvider gateways.AuthTokenProvider,
	signUpCodeTTL time.Duration,
	authTokenTTL time.Duration) *AuthService {
	return &AuthService{usersRepo, signUpCodeRepo, signUpCodeTTL, authTokenProvider, authTokenTTL}
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
	return authToken, user, nil
}

package auth

import (
	"context"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways"
)

type AuthService struct {
	usersRepo gateways.UsersRepo
}

func New(usersRepo gateways.UsersRepo) *AuthService {
	return &AuthService{usersRepo}
}

func (s *AuthService) SignUp(ctx context.Context, user entity.User, otp string) {

}

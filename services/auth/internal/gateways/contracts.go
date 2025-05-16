// Package repo implements application outer layer logic. Each logic group in own file.
package gateways

import (
	"context"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
)

//go:generate mockgen -source=contracts.go -destination=../../tests/mocks/mocks_gateways.go -package=mocks
type (
	UsersRepo interface {
		Insert(ctx context.Context, user *entity.User) (*entity.User, error)
	}
	SignUpCodeRepo interface {
		GetByEmail(ctx context.Context, email string) (*entity.SignUpCode, error)
	}
	TotpProvider interface {
		Generate() string
	}
	AuthTokenProvider interface {
		NewToken(expires time.Duration, claims map[string]any) (string, error)
		ParseClaimsFromToken(token string) (map[string]any, error)
	}
)

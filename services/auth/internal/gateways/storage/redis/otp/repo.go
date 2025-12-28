package otp_repo

import (
	"context"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/pkg/redis"
)

type Repository struct {
	*redis.Redis
}

func (repo *Repository) GetByEmail(ctx context.Context, email string) (*entity.OTP, error) {
	return nil, nil
}

func (repo *Repository) GetByUserId(ctx context.Context, userId int) (*entity.OTP, error) {
	return nil, nil
}

func (repo *Repository) Delete(ctx context.Context, otp *entity.OTP) error {
	return nil
}

func (repo *Repository) CreateWithTTL(ctx context.Context, otp *entity.OTP, ttl time.Duration) error {
	return nil
}

package redisrepos

import (
	"context"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/pkg/redis"
)

type SessionsRepo struct {
	*redis.Redis
}

func (repo *SessionsRepo) CreateWithTTL(ctx context.Context, session *entity.AuthSession, ttl time.Duration) (*entity.AuthSession, error) {
	return nil, nil
}

func (repo *SessionsRepo) DeleteById(ctx context.Context, id string) error {
	return nil
}

func (repo *SessionsRepo) GetAllByUserId(ctx context.Context, userId int) ([]entity.AuthSession, error) {
	return nil, nil
}

func (repo *SessionsRepo) GetByLoginData(ctx context.Context, ip string, deviceInfo string, userId int) (*entity.AuthSession, error) {
	return nil, nil
}

func (repo *SessionsRepo) GetById(ctx context.Context, id string) (*entity.AuthSession, error) {
	return nil, nil
}

func (repo *SessionsRepo) UpdateById(ctx context.Context, sessionId string, lastSeenAt time.Time, ttl time.Duration) (*entity.AuthSession, error) {
	return nil, nil
}

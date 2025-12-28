package sessions_repo

import (
	"context"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/pkg/redis"
)

type Repository struct {
	*redis.Redis
}

func (repo *Repository) CreateWithTTL(ctx context.Context, session *entity.AuthSession, ttl time.Duration) (*entity.AuthSession, error) {
	return nil, nil
}

func (repo *Repository) DeleteById(ctx context.Context, id string) error {
	return nil
}

func (repo *Repository) GetAllByUserId(ctx context.Context, userId int) ([]entity.AuthSession, error) {
	return nil, nil
}

func (repo *Repository) GetByLoginData(ctx context.Context, ip string, deviceInfo string, userId int) (*entity.AuthSession, error) {
	return nil, nil
}

func (repo *Repository) GetById(ctx context.Context, id string) (*entity.AuthSession, error) {
	return nil, nil
}

func (repo *Repository) UpdateById(ctx context.Context, sessionId string, lastSeenAt time.Time, ttl time.Duration) (*entity.AuthSession, error) {
	return nil, nil
}

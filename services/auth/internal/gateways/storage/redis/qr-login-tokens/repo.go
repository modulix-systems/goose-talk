package qrlogintokens_repo

import (
	"context"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/pkg/redis"
)

type Repository struct {
	*redis.Redis
}

func (repo *Repository) CreateWithTTL(ctx context.Context, token *entity.QRCodeLoginToken, ttl time.Duration) (*entity.QRCodeLoginToken, error) {
	return nil, nil
}

func (repo *Repository) GetByValue(ctx context.Context, val string) (*entity.QRCodeLoginToken, error) {
	return nil, nil
}

func (repo *Repository) DeleteAllByClientId(ctx context.Context, sessionId string) error {
	return nil
}

func (repo *Repository) DeleteByValue(ctx context.Context, val string) error {
	return nil
}

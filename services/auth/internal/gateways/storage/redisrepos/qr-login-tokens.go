package redisrepos

import (
	"context"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/pkg/redis"
)

type QRLoginTokensRepo struct {
	*redis.Redis
}

func (repo *QRLoginTokensRepo) CreateWithTTL(ctx context.Context, token *entity.QRCodeLoginToken, ttl time.Duration) (*entity.QRCodeLoginToken, error) {
	return nil, nil
}

func (repo *QRLoginTokensRepo) GetByValue(ctx context.Context, val string) (*entity.QRCodeLoginToken, error) {
	return nil, nil
}

func (repo *QRLoginTokensRepo) DeleteAllByClientId(ctx context.Context, sessionId string) error {
	return nil
}

func (repo *QRLoginTokensRepo) DeleteByValue(ctx context.Context, val string) error {
	return nil
}

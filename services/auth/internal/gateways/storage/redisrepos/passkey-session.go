package redisrepos

import (
	"context"
	"encoding/json"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/pkg/redis"
)

type PasskeySessionsRepo struct {
	*redis.Redis
}


func (repo *PasskeySessionsRepo) Create(ctx context.Context, session *entity.PasskeyRegistrationSession) error {
	serializedPasskeySession, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return repo.Set(ctx, prefixPasskeySession(session.UserId), serializedPasskeySession, 10*time.Minute).Err()
}

func (repo *PasskeySessionsRepo) GetByUserId(ctx context.Context, userId int) (*entity.PasskeyRegistrationSession, error) {
	passkeySessionJson, err := repo.Get(ctx, prefixPasskeySession(userId)).Result()
	if err != nil {
		return nil, mapError(err)
	}

	var passkeySession entity.PasskeyRegistrationSession
	if err := json.Unmarshal([]byte(passkeySessionJson), &passkeySession); err != nil {
		return nil, err
	}

	return &passkeySession, nil
}

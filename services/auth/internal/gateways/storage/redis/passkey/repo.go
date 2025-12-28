package passkey_repo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/pkg/redis"
	goredis "github.com/redis/go-redis/v9"
)

type Repository struct {
	*redis.Redis
}

func prefixKey(userId int) string {
	return fmt.Sprintf("passkey_session:%d", userId)
}

func (repo *Repository) Create(ctx context.Context, session *entity.PasskeyRegistrationSession) error {
	serializedPasskeySession, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return repo.Set(ctx, prefixKey(session.UserId), serializedPasskeySession, 10*time.Minute).Err()
}

func (repo *Repository) GetByUserId(ctx context.Context, userId int) (*entity.PasskeyRegistrationSession, error) {
	passkeySessionJson, err := repo.Get(ctx, prefixKey(userId)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	var passkeySession entity.PasskeyRegistrationSession
	if err := json.Unmarshal([]byte(passkeySessionJson), &passkeySession); err != nil {
		return nil, err
	}

	return &passkeySession, nil
}

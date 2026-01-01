package redisrepos

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/pkg/redis"
)

type QRLoginTokensRepo struct {
	*redis.Redis
}

type TokenData struct {
	IPAddr string
	DeviceInfo string
}

func (repo *QRLoginTokensRepo) CreateWithTTL(ctx context.Context, token *entity.QRCodeLoginToken, ttl time.Duration) error {
	data := TokenData{IPAddr: token.IPAddr, DeviceInfo: token.DeviceInfo}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("redisrepos - QRLoginTokenRepo.CreateWithTTL - json.Marshal: %w", err)
	}
	if err := repo.Set(ctx, prefixQRLoginToken(token.Value, token.ClientId), string(jsonData), ttl).Err(); err != nil {
		return mapError(err)
	}
	return nil
}

func (repo *QRLoginTokensRepo) FindOne(ctx context.Context, value string, clientId string) (*entity.QRCodeLoginToken, error) {
	jsonData, err := repo.Get(ctx, prefixQRLoginToken(value, clientId)).Result()
	if err != nil {
		return nil, mapError(err)
	}
	var data TokenData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("redisrepos - QRLoginTokenRepo.GetByValue - json.Unmarshal: %w", err)
	}
	return &entity.QRCodeLoginToken{Value: value, ClientId: clientId, IPAddr: data.IPAddr, DeviceInfo: data.DeviceInfo}, nil
}

func (repo *QRLoginTokensRepo) DeleteAllByClient(ctx context.Context, clientId string) error {
	keys, err := repo.Keys(ctx, prefixQRLoginTokenSearchByClient(clientId)).Result()
	if err != nil {
		return mapError(err)
	}
	if err := repo.Del(ctx, keys...).Err(); err != nil {
		return mapError(err)
	}
	return nil
}

package redisrepos

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/utils"
	"github.com/modulix-systems/goose-talk/pkg/redis"
)

type AuthSessionsRepo struct {
	*redis.Redis
}

type SessionData struct {
	LastSeenAt  time.Time        `redis:"LastSeenAt"`
	CreatedAt   time.Time        `redis:"CreatedAt"`
	IsLongLived utils.BoolString `redis:"IsLongLived"`
	Location    string           `redis:"Location"`
	IPAddr      string           `redis:"IPAddr"`
	DeviceInfo  string           `redis:"DeviceInfo"`
}

func (repo *AuthSessionsRepo) CreateWithTTL(ctx context.Context, session *entity.AuthSession, ttl time.Duration) (*entity.AuthSession, error) {
	newSession := *session

	now := time.Now().UTC()
	if newSession.CreatedAt.IsZero() {
		newSession.CreatedAt = now
	}
	if newSession.LastSeenAt.IsZero() {
		newSession.LastSeenAt = now
	}
	data := SessionData{
		LastSeenAt:  newSession.LastSeenAt,
		CreatedAt:   newSession.CreatedAt,
		IsLongLived: utils.BoolString(newSession.IsLongLived),
		Location:    newSession.Location,
		IPAddr:      newSession.IPAddr,
		DeviceInfo:  newSession.DeviceInfo,
	}

	key := prefixAuthSession(session.UserId, session.ID)

	if err := repo.HSet(ctx, key, data).Err(); err != nil {
		return nil, mapError(err)
	}

	if err := repo.Expire(ctx, key, ttl).Err(); err != nil {
		return nil, mapError(err)
	}

	return &newSession, nil
}

func (repo *AuthSessionsRepo) GetAllByUserId(ctx context.Context, userId int) ([]entity.AuthSession, error) {
	keys, err := repo.Keys(ctx, prefixAuthSessionSearchByUser(userId)).Result()
	if err != nil {
		return nil, mapError(fmt.Errorf("redisrepos - AuthSessionsRepo.GetAllByUserId - repo.Keys(%s): %w", prefixAuthSessionSearchByUser(userId), err))
	}

	sessions := make([]entity.AuthSession, 0, len(keys))

	for _, key := range keys {
		session, err := repo.getSession(ctx, key)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, *session)
	}

	return sessions, nil
}

func (repo *AuthSessionsRepo) GetByLoginData(ctx context.Context, userId int, ip string, deviceInfo string) (*entity.AuthSession, error) {
	sessions, err := repo.GetAllByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	deviceInfo = strings.ToLower(deviceInfo)
	for _, session := range sessions {
		if session.IPAddr == ip && strings.ToLower(session.DeviceInfo) == deviceInfo {
			return &session, nil
		}
	}

	return nil, storage.ErrNotFound
}

func (repo *AuthSessionsRepo) getSession(ctx context.Context, key string) (*entity.AuthSession, error) {
	rawData, err := repo.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redisrepos - AuthSessionsRepo.GetAllByUserId - repo.HGetAll(%s): %w", key, err)
	}
	if len(rawData) == 0 {
		return nil, storage.ErrNotFound
	}
	jsonRawData, err := json.Marshal(rawData)
	if err != nil {
		return nil, fmt.Errorf("redisrepos - AuthSessionsRepo.GetAllByUserId - json.Marshal(%s): %w", rawData, err)
	}

	var sessionData SessionData
	if err := json.Unmarshal(jsonRawData, &sessionData); err != nil {
		return nil, fmt.Errorf("redisrepos - AuthSessionsRepo.GetAllByUserId - json.Unmarshal(%s): %w", jsonRawData, err)
	}

	return &entity.AuthSession{
		ID:          extractAuthSessionId(key),
		UserId:      extractAuthSessionUserId(key),
		LastSeenAt:  sessionData.LastSeenAt,
		CreatedAt:   sessionData.CreatedAt,
		IsLongLived: bool(sessionData.IsLongLived),
		Location:    sessionData.Location,
		IPAddr:      sessionData.IPAddr,
		DeviceInfo:  sessionData.DeviceInfo,
	}, nil
}

func (repo *AuthSessionsRepo) GetById(ctx context.Context, userId int, sessionId string) (*entity.AuthSession, error) {
	return repo.getSession(ctx, prefixAuthSession(userId, sessionId))
}

func (repo *AuthSessionsRepo) UpdateById(ctx context.Context, userId int, sessionId string, lastSeenAt time.Time, ttl time.Duration) error {
	if !lastSeenAt.IsZero() {
		if err := repo.HSet(ctx, prefixAuthSession(userId, sessionId), "LastSeenAt", lastSeenAt).Err(); err != nil {
			return mapError(err)
		}
	}

	if ttl != 0 {
		if err := repo.Expire(ctx, prefixAuthSession(userId, sessionId), ttl).Err(); err != nil {
			return mapError(err)
		}
	}

	return nil
}

func (repo *AuthSessionsRepo) DeleteById(ctx context.Context, userId int, sessionId string) error {
	if err := repo.Del(ctx, prefixAuthSession(userId, sessionId)).Err(); err != nil {
		return mapError(err)
	}
	return nil
}

func (repo *AuthSessionsRepo) DeleteAllByUserId(ctx context.Context, userId int, excludeSessionId string) error {
	keys, err := repo.Keys(ctx, prefixAuthSessionSearchByUser(userId)).Result()
	if err != nil {
		return mapError(fmt.Errorf("redisrepos - AuthSessionsRepo.DeleteAllByUserId - repo.Keys(%s): %w", prefixAuthSessionSearchByUser(userId), err))
	}
	if excludeSessionId != "" {
		excludeIndex := slices.IndexFunc(keys, func(key string) bool {
			return extractAuthSessionId(key) == excludeSessionId
		})
		if excludeIndex != -1 {
			keys = slices.Delete(keys, excludeIndex, excludeIndex+1)
		}
	}
	if err := repo.Del(ctx, keys...).Err(); err != nil {
		return mapError(err)
	}
	return nil
}

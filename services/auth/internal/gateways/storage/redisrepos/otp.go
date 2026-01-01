package redisrepos

import (
	"context"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/pkg/redis"
)

type OtpRepo struct {
	*redis.Redis
}

func (repo *OtpRepo) GetByEmail(ctx context.Context, email string) (*entity.OTP, error) {
	otpCode, err := repo.Get(ctx, prefixOtpByEmail(email)).Result()
	if err != nil {
		return nil, mapError(err)
	}
	return &entity.OTP{Code: []byte(otpCode), UserEmail: email}, nil
}

func (repo *OtpRepo) GetByUserId(ctx context.Context, userId int) (*entity.OTP, error) {
	otpCode, err := repo.Get(ctx, prefixOtpByUserId(userId)).Result()
	if err != nil {
		return nil, mapError(err)
	}
	return &entity.OTP{Code: []byte(otpCode), UserId: userId}, nil
}

func (repo *OtpRepo) Delete(ctx context.Context, otp *entity.OTP) error {
	if err := repo.Del(ctx, repo.GetKey(otp)).Err(); err != nil {
		return mapError(err)
	}
	return nil
}

func (repo *OtpRepo) CreateWithTTL(ctx context.Context, otp *entity.OTP, ttl time.Duration) error {
	if err := repo.Set(ctx, repo.GetKey(otp), otp.Code, ttl).Err(); err != nil {
		return mapError(err)
	}
	return nil
}

func (repo *OtpRepo) GetKey(otp *entity.OTP) string {
	if otp.UserId != 0 {
		return prefixOtpByUserId(otp.UserId)
	} 
	return prefixOtpByEmail(otp.UserEmail)
}
package redisrepos_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage/redisrepos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOtpWithTTL(t *testing.T) {
	testSuite := redisrepos.NewTestSuite(t)
	ctx := context.Background()
	t.Run("with email", func(t *testing.T) {
		mockOtp := &entity.OTP{
			Code:      []byte(gofakeit.Numerify("######")),
			UserEmail: gofakeit.Email(),
		}
		expectedTTL := time.Minute

		err := testSuite.Otp.CreateWithTTL(ctx, mockOtp, expectedTTL)
		actualTTL, err := testSuite.RedisClient.TTL(ctx, testSuite.Otp.GetKey(mockOtp)).Result()
		assert.NoError(t, err)
		assert.Equal(t, expectedTTL, actualTTL)
		assert.NoError(t, err)
		otp, err := testSuite.Otp.GetByEmail(ctx, mockOtp.UserEmail)
		require.NoError(t, err)
		assert.Equal(t, mockOtp.Code, otp.Code)
		assert.Equal(t, mockOtp.UserEmail, otp.UserEmail)
	})

	t.Run("with user id", func(t *testing.T) {
		mockOtp := &entity.OTP{
			Code:      []byte(gofakeit.Numerify("######")),
			UserId: 1,
		}
		expectedTTL := time.Minute

		err := testSuite.Otp.CreateWithTTL(ctx, mockOtp, expectedTTL)
		actualTTL, err := testSuite.RedisClient.TTL(ctx, testSuite.Otp.GetKey(mockOtp)).Result()
		assert.NoError(t, err)
		assert.Equal(t, expectedTTL, actualTTL)
		assert.NoError(t, err)
		otp, err := testSuite.Otp.GetByUserId(ctx, mockOtp.UserId)
		require.NoError(t, err)
		assert.Equal(t, mockOtp.Code, otp.Code)
		assert.Equal(t, mockOtp.UserId, otp.UserId)
	})
}

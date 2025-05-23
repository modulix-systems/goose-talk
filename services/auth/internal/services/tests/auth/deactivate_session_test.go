package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDeactivateSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockSessionId := gofakeit.Number(1, 1000)
	mockUserId := gofakeit.Number(1, 1000)
	t.Run("success", func(t *testing.T) {
		authSuite.mockSessionsRepo.EXPECT().UpdateForUserById(
			ctx, mockUserId, mockSessionId, gomock.Any(),
		).DoAndReturn(func(ctx context.Context, userId int, sessionId int, ts time.Time) error {
			assert.WithinDuration(t, time.Now(), ts, time.Second)
			return nil
		})

		err := authSuite.service.DeactivateSession(ctx, mockUserId, mockSessionId)
		assert.NoError(t, err)
	})
	t.Run("not found", func(t *testing.T) {
		authSuite.mockSessionsRepo.EXPECT().UpdateForUserById(
			ctx, mockUserId, mockSessionId, gomock.Any(),
		).DoAndReturn(func(ctx context.Context, userId int, sessionId int, deactivatedAt time.Time) error {
			assert.WithinDuration(t, time.Now(), deactivatedAt, time.Second)
			return storage.ErrNotFound
		})

		err := authSuite.service.DeactivateSession(ctx, mockUserId, mockSessionId)
		assert.ErrorIs(t, err, auth.ErrSessionNotFound)
	})
}

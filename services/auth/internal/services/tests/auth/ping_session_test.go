package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/schemas"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPingSessionSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockSession := helpers.MockUserSession(true)
	expectedSessionId := mockSession.ID
	actAndAssert := func() {
		session, err := authSuite.service.PingSession(ctx, expectedSessionId)
		require.NotNil(t, session)
		assert.Equal(t, expectedSessionId, session.ID)
		assert.NoError(t, err)
	}
	t.Run("with expiry update", func(t *testing.T) {
		mockSession.ExpiresAt = time.Now().Add(authSuite.service.SessionTTLThreshold)
		authSuite.mockSessionsRepo.EXPECT().GetById(ctx, expectedSessionId).Return(mockSession, nil)
		authSuite.mockSessionsRepo.EXPECT().UpdateById(
			ctx, mockSession.ID, gomock.Any()).
			DoAndReturn(func(ctx context.Context, sessionId string, payload *schemas.SessionUpdatePayload) (*entity.AuthSession, error) {
				assert.WithinDuration(t, payload.LastSeenAt, time.Now(), time.Second)
				assert.Equal(t, mockSession.ExpiresAt.Add(authSuite.service.SessionTTLAddend), payload.ExpiresAt)
				assert.Empty(t, payload.DeactivatedAt)
				return mockSession, nil
			})
		actAndAssert()
	})
	t.Run("without updating expiry", func(t *testing.T) {
		mockSession.ExpiresAt = time.Now().Add(100 * time.Hour)
		authSuite.mockSessionsRepo.EXPECT().GetById(ctx, expectedSessionId).Return(mockSession, nil)
		authSuite.mockSessionsRepo.EXPECT().UpdateById(
			ctx, mockSession.ID, gomock.Any()).
			DoAndReturn(func(ctx context.Context, sessionId string, payload *schemas.SessionUpdatePayload) (*entity.AuthSession, error) {
				assert.WithinDuration(t, payload.LastSeenAt, time.Now(), time.Second)
				assert.Empty(t, payload.DeactivatedAt)
				return mockSession, nil
			})
		actAndAssert()
	})
}

func TestPingSessionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	fakeSessionId := gofakeit.UUID()
	authSuite.mockSessionsRepo.EXPECT().GetById(ctx, fakeSessionId).Return(nil, storage.ErrNotFound)

	session, err := authSuite.service.PingSession(ctx, fakeSessionId)
	assert.Empty(t, session)
	assert.ErrorIs(t, err, auth.ErrSessionNotFound)
}

func TestPingSessionNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockToken := gofakeit.UUID()
	mockSession := helpers.MockUserSession(false)
	authSuite.mockSessionsRepo.EXPECT().GetById(ctx, mockToken).Return(mockSession, nil)

	session, err := authSuite.service.PingSession(ctx, mockToken)
	assert.Empty(t, session)
	assert.ErrorIs(t, err, auth.ErrSessionNotFound)
}

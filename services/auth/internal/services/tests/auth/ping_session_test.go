package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways"
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
	mockToken := gofakeit.UUID()
	mockSession := helpers.MockUserSession(true)
	mockSession.AccessToken = mockToken
	actAndAssert := func() {
		session, err := authSuite.service.PingSession(ctx, mockToken)
		require.NotNil(t, session)
		assert.Equal(t, session.ID, mockSession.ID)
		assert.NoError(t, err)
	}
	t.Run("success with valid token", func(t *testing.T) {
		authSuite.mockAuthTokenProvider.EXPECT().
			ParseClaimsFromToken(mockToken).
			Return(map[string]any{"uid": "123"}, nil)
		authSuite.mockSessionsRepo.EXPECT().GetByToken(ctx, mockToken).Return(mockSession, nil)
		authSuite.mockSessionsRepo.EXPECT().UpdateById(
			ctx, mockSession.ID, gomock.Any()).
			DoAndReturn(func(ctx context.Context, sessionId int, payload *schemas.SessionUpdatePayload) (*entity.UserSession, error) {
				assert.WithinDuration(t, payload.LastSeenAt, time.Now(), time.Second)
				assert.Empty(t, payload.DeactivatedAt)
				assert.Empty(t, payload.AccessToken)
				return mockSession, nil
			})
		actAndAssert()
	})
	t.Run("success with expired token", func(t *testing.T) {
		authSuite.mockAuthTokenProvider.EXPECT().
			ParseClaimsFromToken(mockToken).
			Return(nil, gateways.ErrExpiredToken)
		authSuite.mockSessionsRepo.EXPECT().GetByToken(ctx, mockToken).Return(mockSession, nil)
		authSuite.mockSessionsRepo.EXPECT().UpdateById(
			ctx, mockSession.ID, gomock.Any()).
			DoAndReturn(func(ctx context.Context, sessionId int, payload *schemas.SessionUpdatePayload) (*entity.UserSession, error) {
				assert.NotNil(t, payload.DeactivatedAt)
				assert.WithinDuration(t, *payload.DeactivatedAt, time.Now(), time.Second)
				assert.Empty(t, payload.LastSeenAt)
				assert.Empty(t, payload.AccessToken)
				return mockSession, nil
			})
		actAndAssert()
	})
}

func TestPingSessionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockToken := gofakeit.UUID()
	authSuite.mockSessionsRepo.EXPECT().GetByToken(ctx, mockToken).Return(nil, storage.ErrNotFound)

	session, err := authSuite.service.PingSession(ctx, mockToken)
	assert.Empty(t, session)
	assert.ErrorIs(t, err, auth.ErrSessionNotFound)
}

func TestPingSessionNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockToken := gofakeit.UUID()
	mockSession := helpers.MockUserSession(false)
	mockSession.AccessToken = mockToken
	authSuite.mockSessionsRepo.EXPECT().GetByToken(ctx, mockToken).Return(mockSession, nil)

	session, err := authSuite.service.PingSession(ctx, mockToken)
	assert.Empty(t, session)
	assert.ErrorIs(t, err, auth.ErrSessionNotFound)
}

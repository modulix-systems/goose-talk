package auth_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetCurrentSessionSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockToken := gofakeit.UUID()
	mockSession := helpers.MockUserSession(true)
	mockSession.AccessToken = mockToken
	authSuite.mockSessionsRepo.EXPECT().GetByToken(ctx, mockToken).Return(mockSession, nil)

	session, err := authSuite.service.GetCurrentSession(ctx, mockToken)
	require.NotNil(t, session)
	assert.Equal(t, session.ID, mockSession.ID)
	assert.Equal(t, session.AccessToken, mockSession.AccessToken)
	assert.NoError(t, err)
}

func TestGetCurrentSessionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockToken := gofakeit.UUID()
	authSuite.mockSessionsRepo.EXPECT().GetByToken(ctx, mockToken).Return(nil, storage.ErrNotFound)

	session, err := authSuite.service.GetCurrentSession(ctx, mockToken)
	assert.Empty(t, session)
	assert.ErrorIs(t, err, auth.ErrSessionNotFound)
}

func TestGetCurrentSessionNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockToken := gofakeit.UUID()
	mockSession := helpers.MockUserSession(false)
	mockSession.AccessToken = mockToken
	authSuite.mockSessionsRepo.EXPECT().GetByToken(ctx, mockToken).Return(nil, storage.ErrNotFound)

	session, err := authSuite.service.GetCurrentSession(ctx, mockToken)
	assert.Empty(t, session)
	assert.ErrorIs(t, err, auth.ErrSessionNotFound)
}

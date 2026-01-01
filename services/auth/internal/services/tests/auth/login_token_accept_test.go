package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAcceptLoginTokenSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockUser := helpers.MockUser()
	mockSession := helpers.MockAuthSession(true)
	mockSession.UserId = mockUser.ID
	mockLoginToken := helpers.MockLoginToken(authSuite.mockTTL)
	mockSession.ClientIdentity = mockLoginToken.ClientIdentity
	authSuite.mockUsersRepo.EXPECT().GetByID(ctx, mockUser.ID).Return(mockUser, nil)
	authSuite.mockLoginTokenRepo.EXPECT().GetByValue(ctx, mockLoginToken.Value).Return(mockLoginToken, nil)
	authSuite.mockLoginTokenRepo.EXPECT().UpdateAuthSessionByClientId(ctx, mockLoginToken.ClientId, mockSession.ID).Return(nil)
	authSuite.mockSecurityProvider.EXPECT().GenerateSessionId().Return(mockSession.ID)
	setAuthSessionExpectations(t, ctx, authSuite, mockUser, mockSession, gofakeit.Bool(), false)

	session, err := authSuite.service.AcceptLoginToken(ctx, mockUser.ID, mockLoginToken.Value)

	require.NoError(t, err)
	assert.Equal(t, mockSession.ID, session.ID)
	assert.True(t, session.IsActive())
}

func TestAcceptLoginTokenNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockLoginToken := helpers.MockLoginToken(authSuite.mockTTL)
	authSuite.mockLoginTokenRepo.EXPECT().GetByValue(ctx, mockLoginToken.Value).Return(nil, storage.ErrNotFound)

	session, err := authSuite.service.AcceptLoginToken(ctx, gofakeit.Number(1, 1000), mockLoginToken.Value)

	assert.Empty(t, session)
	assert.ErrorIs(t, err, auth.ErrInvalidLoginToken)
}

func TestAcceptLoginTokenUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockLoginToken := helpers.MockLoginToken(authSuite.mockTTL)
	mockUser := helpers.MockUser()
	authSuite.mockUsersRepo.EXPECT().GetByID(ctx, mockUser.ID).Return(nil, storage.ErrNotFound)
	authSuite.mockLoginTokenRepo.EXPECT().GetByValue(ctx, mockLoginToken.Value).Return(mockLoginToken, nil)

	session, err := authSuite.service.AcceptLoginToken(ctx, mockUser.ID, mockLoginToken.Value)

	assert.Empty(t, session)
	assert.ErrorIs(t, err, auth.ErrUserNotFound)
}

func TestAcceptLoginTokenExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockUser := helpers.MockUser()
	mockSession := helpers.MockAuthSession(true)
	mockLoginToken := helpers.MockLoginToken(authSuite.mockTTL)
	mockLoginToken.ExpiresAt = time.Now()
	mockSession.ClientIdentity = mockLoginToken.ClientIdentity
	authSuite.mockLoginTokenRepo.EXPECT().GetByValue(ctx, mockLoginToken.Value).Return(mockLoginToken, nil)

	session, err := authSuite.service.AcceptLoginToken(ctx, mockUser.ID, mockLoginToken.Value)

	assert.Empty(t, session)
	assert.ErrorIs(t, err, auth.ErrExpiredLoginToken)
}

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
	mockSession := helpers.MockUserSession(true)
	mockSession.UserId = mockUser.ID
	mockLoginToken := helpers.MockLoginToken(authSuite.tokenTTL)
	mockSession.ClientIdentity = mockLoginToken.ClientIdentity
	authSuite.mockUsersRepo.EXPECT().GetByID(ctx, mockUser.ID).Return(mockUser, nil)
	authSuite.mockLoginTokenRepo.EXPECT().GetByValue(ctx, mockLoginToken.Val).Return(mockLoginToken, nil)
	authSuite.mockLoginTokenRepo.EXPECT().UpdateAuthSessionByClientId(ctx, mockLoginToken.ClientId, mockSession.ID).Return(nil)
	authSuite.mockAuthTokenProvider.EXPECT().NewToken(authSuite.tokenTTL, map[string]any{"uid": mockUser.ID}).Return(mockSession.AccessToken, nil)
	setAuthSessionExpectations(t, ctx, authSuite, mockUser, mockSession, gofakeit.Bool(), false)

	session, err := authSuite.service.AcceptLoginToken(ctx, mockUser.ID, mockLoginToken.Val)

	require.NoError(t, err)
	assert.Equal(t, mockSession.ID, session.ID)
	assert.True(t, session.IsActive())
	assert.Equal(t, mockSession.AccessToken, session.AccessToken)
}

func TestAcceptLoginTokenNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockLoginToken := helpers.MockLoginToken(authSuite.tokenTTL)
	authSuite.mockLoginTokenRepo.EXPECT().GetByValue(ctx, mockLoginToken.Val).Return(nil, storage.ErrNotFound)

	session, err := authSuite.service.AcceptLoginToken(ctx, gofakeit.Number(1, 1000), mockLoginToken.Val)

	assert.Empty(t, session)
	assert.ErrorIs(t, err, auth.ErrInvalidLoginToken)
}

func TestAcceptLoginTokenUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockLoginToken := helpers.MockLoginToken(authSuite.tokenTTL)
	mockUser := helpers.MockUser()
	authSuite.mockUsersRepo.EXPECT().GetByID(ctx, mockUser.ID).Return(nil, storage.ErrNotFound)
	authSuite.mockLoginTokenRepo.EXPECT().GetByValue(ctx, mockLoginToken.Val).Return(mockLoginToken, nil)

	session, err := authSuite.service.AcceptLoginToken(ctx, mockUser.ID, mockLoginToken.Val)

	assert.Empty(t, session)
	assert.ErrorIs(t, err, auth.ErrUserNotFound)
}

func TestAcceptLoginTokenExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockUser := helpers.MockUser()
	mockSession := helpers.MockUserSession(true)
	mockLoginToken := helpers.MockLoginToken(authSuite.tokenTTL)
	mockLoginToken.ExpiresAt = time.Now()
	mockSession.ClientIdentity = mockLoginToken.ClientIdentity
	authSuite.mockLoginTokenRepo.EXPECT().GetByValue(ctx, mockLoginToken.Val).Return(mockLoginToken, nil)

	session, err := authSuite.service.AcceptLoginToken(ctx, mockUser.ID, mockLoginToken.Val)

	assert.Empty(t, session)
	assert.ErrorIs(t, err, auth.ErrExpiredLoginToken)
}

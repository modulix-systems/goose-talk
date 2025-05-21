package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetActiveSessionsSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	mockSessions := []entity.UserSession{
		*helpers.MockUserSession(true),
		*helpers.MockUserSession(true),
	}
	mockUserId := gofakeit.Number(1, 1000)
	ctx := context.Background()
	mockAuthToken := gofakeit.UUID()
	authSuite.mockAuthTokenProvider.EXPECT().
		ParseClaimsFromToken(mockAuthToken).
		Return(map[string]any{"uid": mockUserId}, nil)
	authSuite.mockSessionsRepo.EXPECT().
		GetAllForUser(ctx, mockUserId, true).
		Return(mockSessions, nil)
	sessions, err := authSuite.service.GetActiveSessions(ctx, mockAuthToken)
	assert.NoError(t, err)
	assert.Equal(t, sessions, mockSessions)
}

func TestGetActiveSessionsExpiredToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockAuthToken := gofakeit.UUID()
	authSuite.mockAuthTokenProvider.EXPECT().
		ParseClaimsFromToken(mockAuthToken).
		Return(nil, gateways.ErrExpiredToken)
	sessions, err := authSuite.service.GetActiveSessions(ctx, mockAuthToken)
	assert.ErrorIs(t, err, auth.ErrExpiredAuthToken)
	assert.Empty(t, sessions)
}

func TestGetActiveSessionsInvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockAuthToken := gofakeit.UUID()
	authSuite.mockAuthTokenProvider.EXPECT().ParseClaimsFromToken(mockAuthToken).
		Return(nil, errors.New("malformed token, invalid sig, etc..."))
	sessions, err := authSuite.service.GetActiveSessions(ctx, mockAuthToken)
	assert.ErrorIs(t, err, auth.ErrInvalidAuthToken)
	assert.Empty(t, sessions)
}

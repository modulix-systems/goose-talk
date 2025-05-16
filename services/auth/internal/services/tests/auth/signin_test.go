package auth_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/schemas"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/tests/suite"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func mockSignInPayload() *schemas.SignInSchema {
	return &schemas.SignInSchema{
		Login:    gofakeit.Username(),
		Password: suite.RandomPassword(),
	}
}

func TestSignInSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	expectedToken := "testtoken"
	mockUser := &entity.User{ID: gofakeit.Number(1, 1000), Password: []byte("hashedPass")}
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockUser.Password, dto.Password).Return(true, nil)
	authSuite.mockAuthTokenProvider.EXPECT().
		NewToken(authSuite.tokenTTL, map[string]any{"uid": mockUser.ID}).
		Return(expectedToken, nil)

	token, user, err := authSuite.service.SignIn(ctx, dto)

	assert.NoError(t, err)
	assert.Equal(t, token, expectedToken)
	assert.Equal(t, user.ID, mockUser.ID)
}

func TestSignInNotFoundUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(nil, storage.ErrNotFound)

	token, user, err := authSuite.service.SignIn(ctx, dto)

	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	assert.Empty(t, token)
	assert.Empty(t, user)
}

func TestSignInInvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	mockUser := &entity.User{ID: gofakeit.Number(1, 1000), Password: []byte("hashedPass")}
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockUser.Password, dto.Password).Return(false, nil)

	token, user, err := authSuite.service.SignIn(ctx, dto)

	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
	assert.Empty(t, token)
	assert.Empty(t, user)
}

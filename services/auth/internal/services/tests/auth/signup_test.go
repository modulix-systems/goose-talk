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
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func mockSignUpPayload() *schemas.SignUpSchema {
	return &schemas.SignUpSchema{
		Username:         gofakeit.Username(),
		Email:            gofakeit.Email(),
		FirstName:        gofakeit.FirstName(),
		LastName:         gofakeit.LastName(),
		ConfirmationCode: "testcode",
	}
}

func TestSignupSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	suite := NewAuthTestSuite(ctrl)
	dto := mockSignUpPayload()
	expectedToken := "auth token"
	ctx := context.Background()
	expectedUser := entity.User{ID: gofakeit.Number(1, 1000), Email: dto.Email}
	suite.mockAuthTokenProvider.EXPECT().
		NewToken(suite.tokenTTL, map[string]any{"uid": expectedUser.ID}).
		Return(expectedToken, nil)
	suite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(
		&entity.SignUpCode{Code: dto.ConfirmationCode, Email: dto.Email, CreatedAt: time.Now()},
		nil,
	)
	suite.mockUsersRepo.EXPECT().
		Insert(ctx, &entity.User{FirstName: dto.FirstName, LastName: dto.LastName, Email: dto.Email}).
		Return(&expectedUser, nil)

	token, user, err := suite.authService.SignUp(ctx, dto)

	assert.Equal(t, token, expectedToken)
	assert.Equal(t, user.ID, expectedUser.ID)
	assert.NoError(t, err)
}

func TestSignupNotFoundCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	suite := NewAuthTestSuite(ctrl)
	dto := mockSignUpPayload()
	ctx := context.Background()
	suite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(
		nil,
		storage.ErrNotFound,
	)

	token, user, err := suite.authService.SignUp(ctx, dto)

	assert.Empty(t, token)
	assert.Empty(t, user)
	assert.ErrorIs(t, err, auth.ErrInvalidSignUpCode)
}

func TestSignupExpiredCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	suite := NewAuthTestSuite(ctrl)
	dto := mockSignUpPayload()
	ctx := context.Background()
	suite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(
		&entity.SignUpCode{
			Code:      dto.ConfirmationCode,
			Email:     dto.Email,
			CreatedAt: time.Now().Add(-suite.tokenTTL),
		},
		nil,
	)

	token, user, err := suite.authService.SignUp(ctx, dto)

	assert.Empty(t, token)
	assert.Empty(t, user)
	assert.ErrorIs(t, err, auth.ErrExpiredSignUpCode)
}

func TestSignUpUserExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	suite := NewAuthTestSuite(ctrl)
	dto := mockSignUpPayload()
	ctx := context.Background()
	suite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(
		&entity.SignUpCode{
			Code:      dto.ConfirmationCode,
			Email:     dto.Email,
			CreatedAt: time.Now(),
		},
		nil,
	)
	suite.mockUsersRepo.EXPECT().
		Insert(ctx, &entity.User{FirstName: dto.FirstName, LastName: dto.LastName, Email: dto.Email}).
		Return(nil, storage.ErrAlreadyExists)

	token, user, err := suite.authService.SignUp(ctx, dto)

	assert.Empty(t, token)
	assert.Empty(t, user)
	assert.ErrorIs(t, err, auth.ErrUserAlreadyExists)

}

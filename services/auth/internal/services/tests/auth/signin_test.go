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

func TestSignInSuccessNo2FA(t *testing.T) {
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

func TestSignInSuccess2FADisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	expectedToken := "testtoken"
	mockUser := &entity.User{
		ID:       gofakeit.Number(1, 1000),
		Password: []byte("hashedPass"),
		TwoFactorAuth: &entity.TwoFactorAuth{
			DeliveryMethod: entity.TWO_FA_EMAIL,
		},
	}
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

func TestSignInSuccess2FAByUserEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	plainOTPCode := "securetoken"
	hashedOTPCode := []byte(plainOTPCode)
	mockUser := &entity.User{
		ID:       gofakeit.Number(1, 1000),
		Email:    gofakeit.Email(),
		Password: []byte("hashedPass"),
		TwoFactorAuth: &entity.TwoFactorAuth{
			DeliveryMethod: entity.TWO_FA_EMAIL,
			Enabled:        true,
		},
	}
	mockOTP := &entity.OTP{Code: hashedOTPCode, UserEmail: mockUser.Email}
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockUser.Password, dto.Password).Return(true, nil)
	authSuite.mockSecurityProvider.EXPECT().GenerateOTPCode(6).Return(plainOTPCode)
	authSuite.mockSecurityProvider.EXPECT().HashPassword(plainOTPCode).Return(hashedOTPCode, nil)
	authSuite.mockCodeRepo.EXPECT().InsertOrUpdateCode(ctx, mockOTP).Return(nil)
	authSuite.mockMailSender.EXPECT().Send2FAEmail(ctx, mockUser.Email, plainOTPCode).Return(nil)

	token, user, err := authSuite.service.SignIn(ctx, dto)

	assert.NoError(t, err)
	assert.Empty(t, token)
	assert.Equal(t, user.ID, mockUser.ID)
}

func TestSignInSuccess2FAByContactEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	plainOTPCode := "securetoken"
	hashedOTPCode := []byte(plainOTPCode)
	mockUser := &entity.User{
		ID:       gofakeit.Number(1, 1000),
		Email:    gofakeit.Email(),
		Password: []byte("hashedPass"),
		TwoFactorAuth: &entity.TwoFactorAuth{
			DeliveryMethod: entity.TWO_FA_EMAIL,
			Enabled:        true,
			Contact:        gofakeit.Email(),
		},
	}
	mockOTP := &entity.OTP{Code: hashedOTPCode, UserEmail: mockUser.Email}
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockUser.Password, dto.Password).Return(true, nil)
	authSuite.mockSecurityProvider.EXPECT().GenerateOTPCode(6).Return(plainOTPCode)
	authSuite.mockSecurityProvider.EXPECT().HashPassword(plainOTPCode).Return(hashedOTPCode, nil)
	authSuite.mockCodeRepo.EXPECT().InsertOrUpdateCode(ctx, mockOTP).Return(nil)
	authSuite.mockMailSender.EXPECT().Send2FAEmail(ctx, mockUser.TwoFactorAuth.Contact, plainOTPCode).Return(nil)

	token, user, err := authSuite.service.SignIn(ctx, dto)

	assert.NoError(t, err)
	assert.Empty(t, token)
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

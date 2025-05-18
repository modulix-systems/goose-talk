package auth_test

import (
	"context"
	"fmt"
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
	mockUser := suite.MockUser()
	mockUser.TwoFactorAuth = nil
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockUser.Password, dto.Password).Return(true, nil)
	authSuite.mockAuthTokenProvider.EXPECT().
		NewToken(authSuite.tokenTTL, map[string]any{"uid": mockUser.ID}).
		Return(expectedToken, nil)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.NoError(t, err)
	assert.NotNil(t, authInfo)
	assert.Equal(t, authInfo.Token.Val, expectedToken)
	assert.Equal(t, authInfo.Token.Typ, auth.AuthTokenType)
	assert.Equal(t, authInfo.User.ID, mockUser.ID)
}

func TestSignInSuccess2FADisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	expectedToken := "testtoken"
	mockUser := suite.MockUser()
	mockUser.TwoFactorAuth.Enabled = false
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockUser.Password, dto.Password).Return(true, nil)
	authSuite.mockAuthTokenProvider.EXPECT().
		NewToken(authSuite.tokenTTL, map[string]any{"uid": mockUser.ID}).
		Return(expectedToken, nil)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.NoError(t, err)
	assert.Equal(t, authInfo.Token.Val, expectedToken)
	assert.Equal(t, authInfo.Token.Typ, auth.AuthTokenType)
	assert.Equal(t, authInfo.User.ID, mockUser.ID)
}

func TestSignInSuccess2FAByUserEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	plainOTPCode := "securetoken"
	hashedOTPCode := []byte(plainOTPCode)
	mockUser := suite.MockUser()
	mockUser.TwoFactorAuth.Enabled = true
	mockUser.TwoFactorAuth.DeliveryMethod = entity.TWO_FA_EMAIL
	mockUser.TwoFactorAuth.Contact = ""
	mockOTP := &entity.OTP{Code: hashedOTPCode, UserEmail: mockUser.Email}
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockUser.Password, dto.Password).Return(true, nil)
	authSuite.mockSecurityProvider.EXPECT().GenerateOTPCode(6).Return(plainOTPCode)
	authSuite.mockSecurityProvider.EXPECT().HashPassword(plainOTPCode).Return(hashedOTPCode, nil)
	authSuite.mockCodeRepo.EXPECT().InsertOrUpdateCode(ctx, mockOTP).Return(nil)
	authSuite.mockMailSender.EXPECT().Send2FAEmail(ctx, mockUser.Email, plainOTPCode).Return(nil)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.NoError(t, err)
	assert.Empty(t, authInfo.Token)
	assert.Equal(t, authInfo.User.ID, mockUser.ID)
}

func TestSignInSuccess2FAByContactEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	plainOTPCode := "securetoken"
	hashedOTPCode := []byte(plainOTPCode)
	mockUser := suite.MockUser()
	mockUser.TwoFactorAuth.Enabled = true
	mockUser.TwoFactorAuth.DeliveryMethod = entity.TWO_FA_EMAIL
	mockOTP := &entity.OTP{Code: hashedOTPCode, UserEmail: mockUser.Email}
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockUser.Password, dto.Password).Return(true, nil)
	authSuite.mockSecurityProvider.EXPECT().GenerateOTPCode(6).Return(plainOTPCode)
	authSuite.mockSecurityProvider.EXPECT().HashPassword(plainOTPCode).Return(hashedOTPCode, nil)
	authSuite.mockCodeRepo.EXPECT().InsertOrUpdateCode(ctx, mockOTP).Return(nil)
	authSuite.mockMailSender.EXPECT().Send2FAEmail(ctx, mockUser.TwoFactorAuth.Contact, plainOTPCode).Return(nil)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.NoError(t, err)
	assert.Empty(t, authInfo.Token)
	assert.Equal(t, authInfo.User.ID, mockUser.ID)
}

func TestSignInSuccess2FAByContactTG(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	plainOTPCode := "securetoken"
	hashedOTPCode := []byte(plainOTPCode)
	mockUser := suite.MockUser()
	mockUser.TwoFactorAuth.Enabled = true
	mockUser.TwoFactorAuth.DeliveryMethod = entity.TWO_FA_TELEGRAM
	mockOTP := &entity.OTP{Code: hashedOTPCode, UserEmail: mockUser.Email}
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockUser.Password, dto.Password).Return(true, nil)
	authSuite.mockSecurityProvider.EXPECT().HashPassword(plainOTPCode).Return(hashedOTPCode, nil)
	authSuite.mockSecurityProvider.EXPECT().GenerateOTPCode(6).Return(plainOTPCode)
	authSuite.mockCodeRepo.EXPECT().InsertOrUpdateCode(ctx, mockOTP).Return(nil)
	authSuite.mockTgAPI.EXPECT().SendTextMsg(ctx, mockUser.TwoFactorAuth.Contact, fmt.Sprintf("Authorization code: %s", plainOTPCode)).Return(nil)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.NoError(t, err)
	assert.Empty(t, authInfo.Token)
	assert.Equal(t, authInfo.User.ID, mockUser.ID)
}

func TestSignInSuccess2FAByTotp(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	plainOTPCode := "securetoken"
	hashedOTPCode := []byte(plainOTPCode)
	mockUser := suite.MockUser()
	mockUser.TwoFactorAuth.Enabled = true
	mockUser.TwoFactorAuth.DeliveryMethod = entity.TWO_FA_TOTP_APP
	mockOTP := &entity.OTP{Code: hashedOTPCode, UserEmail: mockUser.Email}
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockUser.Password, dto.Password).Return(true, nil)
	authSuite.mockSecurityProvider.EXPECT().GenerateOTPCode(6).Return(plainOTPCode)
	authSuite.mockSecurityProvider.EXPECT().HashPassword(plainOTPCode).Return(hashedOTPCode, nil)
	authSuite.mockCodeRepo.EXPECT().InsertOrUpdateCode(ctx, mockOTP).Return(nil)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.NoError(t, err)
	assert.Equal(t, authInfo.Token.Val, plainOTPCode)
	assert.Equal(t, authInfo.Token.Typ, auth.LoginConfTokenType)
	assert.Equal(t, authInfo.User.ID, mockUser.ID)
}

func TestSignIn2FAUnsupportedMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	plainOTPCode := "securetoken"
	hashedOTPCode := []byte(plainOTPCode)
	const unsupported2FAMethod entity.TwoFADeliveryMethod = -1
	mockUser := suite.MockUser()
	mockUser.TwoFactorAuth.Enabled = true
	mockUser.TwoFactorAuth.DeliveryMethod = unsupported2FAMethod
	mockOTP := &entity.OTP{Code: hashedOTPCode, UserEmail: mockUser.Email}
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockUser.Password, dto.Password).Return(true, nil)
	authSuite.mockSecurityProvider.EXPECT().GenerateOTPCode(6).Return(plainOTPCode)
	authSuite.mockSecurityProvider.EXPECT().HashPassword(plainOTPCode).Return(hashedOTPCode, nil)
	authSuite.mockCodeRepo.EXPECT().InsertOrUpdateCode(ctx, mockOTP).Return(nil)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.Empty(t, authInfo)
	assert.ErrorIs(t, err, auth.ErrUnsupported2FAMethod)
}

func TestSignInNotFoundUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(nil, storage.ErrNotFound)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.Empty(t, authInfo)
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestSignInNotActiveUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	mockUser := suite.MockUser()
	mockUser.IsActive = false
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.Empty(t, authInfo)
	assert.ErrorIs(t, err, auth.ErrDisabledAccount)

}

func TestSignInInvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	mockUser := suite.MockUser()
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockUser.Password, dto.Password).Return(false, nil)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.Empty(t, authInfo)
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)

}

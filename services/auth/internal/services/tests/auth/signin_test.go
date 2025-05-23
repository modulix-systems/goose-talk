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
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func mockSignInPayload() *schemas.SignInSchema {
	return &schemas.SignInSchema{
		Login:    gofakeit.Username(),
		Password: helpers.RandomPassword(),
		ClientIdentitySchema: schemas.ClientIdentitySchema{
			DeviceInfo: gofakeit.UserAgent(),
			IPAddr:     gofakeit.IPv4Address(),
		},
	}
}

func setSignInWith2FAExpectations(
	ctx context.Context,
	authSuite *AuthTestSuite,
	dto *schemas.SignInSchema,
	mockUser *entity.User,
	plainOTPCode string,
) {
	hashedOTPCode := []byte(plainOTPCode)
	mockOTP := &entity.OTP{Code: hashedOTPCode, UserId: mockUser.ID}
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().
		ComparePasswords(mockUser.Password, dto.Password).
		Return(true, nil)
	authSuite.mockSecurityProvider.EXPECT().GenerateOTPCode().Return(plainOTPCode)
	authSuite.mockSecurityProvider.EXPECT().HashPassword(plainOTPCode).Return(hashedOTPCode, nil)
	authSuite.mockCodeRepo.EXPECT().InsertOrUpdateCode(ctx, mockOTP).Return(nil)
}

func TestSignInSuccessNo2FA(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	expectedToken := "testtoken"
	testCases := []struct {
		name          string
		sessionExists bool
		twoFAIncluded bool
	}{
		{
			name:          "2fa is nil",
			sessionExists: true,
			twoFAIncluded: false,
		},
		{
			name:          "2fa disabled",
			sessionExists: true,
			twoFAIncluded: true,
		},
		{
			name:          "2fa is nil new auth session",
			sessionExists: false,
			twoFAIncluded: false,
		},
		{
			name:          "2fa disabled new auth session",
			sessionExists: false,
			twoFAIncluded: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUser := helpers.MockUser()
			mockUser.TwoFactorAuth.Enabled = false
			if !tc.twoFAIncluded {
				mockUser.TwoFactorAuth = nil
			}
			mockSession := helpers.MockUserSession(gofakeit.Bool())
			mockSession.UserId = mockUser.ID
			mockSession.AccessToken = expectedToken
			mockSession.ClientIdentity = &entity.ClientIdentity{DeviceInfo: dto.DeviceInfo, IPAddr: dto.IPAddr}
			authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
			setAuthSessionExpectations(t, ctx, authSuite, mockUser, mockSession, tc.sessionExists)
			authSuite.mockSecurityProvider.EXPECT().
				ComparePasswords(mockUser.Password, dto.Password).
				Return(true, nil)
			authSuite.mockAuthTokenProvider.EXPECT().
				NewToken(authSuite.tokenTTL, map[string]any{"uid": mockUser.ID}).
				Return(expectedToken, nil)

			authInfo, err := authSuite.service.SignIn(ctx, dto)

			assert.NoError(t, err)
			assert.NotNil(t, authInfo)
			assert.Equal(t, authInfo.Session.AccessToken, expectedToken)
			assert.Equal(t, authInfo.User.ID, mockUser.ID)
		})
	}
}

func TestSignInSuccess2FAByUserEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	plainOTPCode := "securetoken"
	mockUser := helpers.MockUser()
	mockUser.TwoFactorAuth.Enabled = true
	mockUser.TwoFactorAuth.DeliveryMethod = entity.TWO_FA_EMAIL
	mockUser.TwoFactorAuth.Contact = ""
	setSignInWith2FAExpectations(ctx, authSuite, dto, mockUser, plainOTPCode)
	authSuite.mockMailSender.EXPECT().Send2FAEmail(ctx, mockUser.Email, plainOTPCode).Return(nil)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.NoError(t, err)
	assert.Empty(t, authInfo.Session)
	assert.Equal(t, authInfo.User.ID, mockUser.ID)
}

func TestSignInSuccess2FAByContactEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	plainOTPCode := "securetoken"
	mockUser := helpers.MockUser()
	mockUser.TwoFactorAuth.Enabled = true
	mockUser.TwoFactorAuth.DeliveryMethod = entity.TWO_FA_EMAIL
	setSignInWith2FAExpectations(ctx, authSuite, dto, mockUser, plainOTPCode)
	authSuite.mockMailSender.EXPECT().
		Send2FAEmail(ctx, mockUser.TwoFactorAuth.Contact, plainOTPCode).
		Return(nil)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.NoError(t, err)
	assert.Empty(t, authInfo.Session)
	assert.Equal(t, authInfo.User.ID, mockUser.ID)
}

func TestSignInSuccess2FAByContactTG(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	plainOTPCode := "securetoken"
	mockUser := helpers.MockUser()
	mockUser.TwoFactorAuth.Enabled = true
	mockUser.TwoFactorAuth.DeliveryMethod = entity.TWO_FA_TELEGRAM
	setSignInWith2FAExpectations(ctx, authSuite, dto, mockUser, plainOTPCode)
	authSuite.mockTgAPI.EXPECT().
		SendTextMsg(ctx, mockUser.TwoFactorAuth.Contact, fmt.Sprintf("Authorization code: %s", plainOTPCode)).
		Return(nil)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.NoError(t, err)
	assert.Empty(t, authInfo.Session)
	assert.Equal(t, authInfo.User.ID, mockUser.ID)
}

func TestSignInSuccess2FAByTotp(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	plainOTPCode := "securetoken"
	mockUser := helpers.MockUser()
	mockUser.TwoFactorAuth.Enabled = true
	mockUser.TwoFactorAuth.DeliveryMethod = entity.TWO_FA_TOTP_APP
	setSignInWith2FAExpectations(ctx, authSuite, dto, mockUser, plainOTPCode)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.NoError(t, err)
	assert.Equal(t, authInfo.SignInConfTokenType, plainOTPCode)
	assert.Equal(t, authInfo.User.ID, mockUser.ID)
}

func TestSignIn2FAUnsupportedMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	dto := mockSignInPayload()
	plainOTPCode := "securetoken"
	const unsupported2FAMethod entity.TwoFADeliveryMethod = -1
	mockUser := helpers.MockUser()
	mockUser.TwoFactorAuth.Enabled = true
	mockUser.TwoFactorAuth.DeliveryMethod = unsupported2FAMethod
	setSignInWith2FAExpectations(ctx, authSuite, dto, mockUser, plainOTPCode)

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
	mockUser := helpers.MockUser()
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
	mockUser := helpers.MockUser()
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Login).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().
		ComparePasswords(mockUser.Password, dto.Password).
		Return(false, nil)

	authInfo, err := authSuite.service.SignIn(ctx, dto)

	assert.Empty(t, authInfo)
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)

}

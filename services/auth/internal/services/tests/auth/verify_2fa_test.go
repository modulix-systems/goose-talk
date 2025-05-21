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
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func mockVerify2FAPayload(twoFaTyp entity.TwoFADeliveryMethod) *schemas.Verify2FASchema {
	schema := &schemas.Verify2FASchema{
		TwoFATyp:   twoFaTyp,
		Code:       "123456",
		Email:      gofakeit.Email(),
		ClientIP:   gofakeit.IPv4Address(),
		DeviceInfo: gofakeit.UserAgent(),
	}
	if twoFaTyp == entity.TWO_FA_TOTP_APP {
		schema.SignInConfToken = "456789"
	}
	return schema
}

func TestVerify2FASuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)

	ctx := context.Background()
	testCases := []struct {
		name          string
		twoFaTyp      entity.TwoFADeliveryMethod
		sessionExists bool
	}{
		{
			name:          "2FA over " + entity.TWO_FA_EMAIL.String(),
			twoFaTyp:      entity.TWO_FA_EMAIL,
			sessionExists: gofakeit.Bool(),
		},
		{
			name:          "2FA over " + entity.TWO_FA_TELEGRAM.String(),
			twoFaTyp:      entity.TWO_FA_TELEGRAM,
			sessionExists: gofakeit.Bool(),
		},
		{
			name:          "2FA over " + entity.TWO_FA_TOTP_APP.String(),
			twoFaTyp:      entity.TWO_FA_TOTP_APP,
			sessionExists: gofakeit.Bool(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dto := mockVerify2FAPayload(tc.twoFaTyp)
			expectedAuthToken := "authtoken"
			mockOTP := helpers.MockOTP()
			mockOTP.UserEmail = dto.Email
			mockUser := helpers.MockUser()
			mockUser.TwoFactorAuth.Enabled = true
			mockSession := helpers.MockUserSession(gofakeit.Bool())
			mockSession.UserId = mockUser.ID
			mockSession.AccessToken = expectedAuthToken
			authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)
			otpToCompare := dto.Code
			if tc.twoFaTyp == entity.TWO_FA_TOTP_APP {
				otpToCompare = dto.SignInConfToken
				authSuite.mockSecurityProvider.EXPECT().
					ValidateTOTP(dto.Code, mockUser.TwoFactorAuth.TotpSecret).
					Return(true)
			}
			authSuite.mockSecurityProvider.EXPECT().
				ComparePasswords(mockOTP.Code, otpToCompare).
				Return(true, nil)
			setAuthSessionExpectations(t, ctx, authSuite, mockUser.ID, mockSession, tc.sessionExists, dto.DeviceInfo, dto.ClientIP, expectedAuthToken)
			authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Email).Return(mockUser, nil)
			authSuite.mockAuthTokenProvider.EXPECT().
				NewToken(authSuite.tokenTTL, map[string]any{"uid": mockUser.ID}).
				Return(expectedAuthToken, nil)
			authToken, err := authSuite.service.Verify2FA(ctx, dto)
			assert.NoError(t, err)
			assert.Equal(t, authToken, expectedAuthToken)
		})
	}
}

func TestVerify2FAOtpNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	for _, twoFaTyp := range entity.OtpDeliveryMethods {
		name := "2FA over " + twoFaTyp.String()
		t.Run(name, func(t *testing.T) {
			dto := mockVerify2FAPayload(twoFaTyp)
			authSuite.mockCodeRepo.EXPECT().
				GetByEmail(ctx, dto.Email).
				Return(nil, storage.ErrNotFound)

			authToken, err := authSuite.service.Verify2FA(ctx, dto)
			assert.ErrorIs(t, err, auth.ErrInvalidOtp)
			assert.Empty(t, authToken)

		})
	}
}

func TestVerify2FAInvalidOTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockOTP := helpers.MockOTP()
	for _, twoFaTyp := range entity.OtpDeliveryMethods {
		name := "2FA over " + twoFaTyp.String()
		t.Run(name, func(t *testing.T) {
			dto := mockVerify2FAPayload(twoFaTyp)
			authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)
			otpToCompare := dto.Code
			if twoFaTyp == entity.TWO_FA_TOTP_APP {
				otpToCompare = dto.SignInConfToken
			}
			authSuite.mockSecurityProvider.EXPECT().
				ComparePasswords(mockOTP.Code, otpToCompare).
				Return(false, nil)

			authToken, err := authSuite.service.Verify2FA(ctx, dto)
			assert.ErrorIs(t, err, auth.ErrInvalidOtp)
			assert.Empty(t, authToken)

		})
	}
}

func TestVerify2FADisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockOTP := helpers.MockOTP()
	mockUser := helpers.MockUser()
	mockUser.TwoFactorAuth.Enabled = false
	for _, twoFaTyp := range entity.OtpDeliveryMethods {
		name := "2FA over " + twoFaTyp.String()
		t.Run(name, func(t *testing.T) {
			dto := mockVerify2FAPayload(twoFaTyp)
			otpToCompare := dto.Code
			if twoFaTyp == entity.TWO_FA_TOTP_APP {
				otpToCompare = dto.SignInConfToken
			}
			authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)
			authSuite.mockSecurityProvider.EXPECT().
				ComparePasswords(mockOTP.Code, otpToCompare).
				Return(true, nil)
			authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Email).Return(mockUser, nil)

			authToken, err := authSuite.service.Verify2FA(ctx, dto)
			assert.ErrorIs(t, err, auth.Err2FANotEnabled)
			assert.Empty(t, authToken)
		})
	}
}

func TestVerify2FATotpAppInvalidTOTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockOTP := helpers.MockOTP()
	mockUser := helpers.MockUser()
	mockUser.TwoFactorAuth.Enabled = true
	dto := mockVerify2FAPayload(entity.TWO_FA_TOTP_APP)
	authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)
	authSuite.mockSecurityProvider.EXPECT().
		ComparePasswords(mockOTP.Code, dto.SignInConfToken).
		Return(true, nil)
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Email).Return(mockUser, nil)
	authSuite.mockSecurityProvider.EXPECT().
		ValidateTOTP(dto.Code, mockUser.TwoFactorAuth.TotpSecret).
		Return(false)

	authToken, err := authSuite.service.Verify2FA(ctx, dto)
	assert.ErrorIs(t, err, auth.ErrInvalidOrExpiredTOTP)
	assert.Empty(t, authToken)
}

func TestVerify2FAOtpExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockOTP := helpers.MockOTP()
	mockOTP.UpdatedAt = time.Now().Add(-authSuite.tokenTTL)
	for _, twoFaTyp := range entity.OtpDeliveryMethods {
		name := "2FA over " + twoFaTyp.String()
		t.Run(name, func(t *testing.T) {
			dto := mockVerify2FAPayload(twoFaTyp)
			authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)

			authToken, err := authSuite.service.Verify2FA(ctx, dto)
			assert.ErrorIs(t, err, auth.ErrOtpExpired)
			assert.Empty(t, authToken)
		})
	}
}

func TestVerify2FAUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockOTP := helpers.MockOTP()
	for _, twoFaTyp := range entity.OtpDeliveryMethods {
		name := "2FA over " + twoFaTyp.String()
		t.Run(name, func(t *testing.T) {
			dto := mockVerify2FAPayload(twoFaTyp)
			authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)
			otpToCompare := dto.Code
			if twoFaTyp == entity.TWO_FA_TOTP_APP {
				otpToCompare = dto.SignInConfToken
			}
			authSuite.mockSecurityProvider.EXPECT().
				ComparePasswords(mockOTP.Code, otpToCompare).
				Return(true, nil)
			authSuite.mockUsersRepo.EXPECT().
				GetByLogin(ctx, dto.Email).
				Return(nil, storage.ErrNotFound)
			authSuite.mockCodeRepo.EXPECT().DeleteByEmail(ctx, dto.Email).Return(nil)

			authToken, err := authSuite.service.Verify2FA(ctx, dto)
			assert.ErrorIs(t, err, auth.ErrUserNotFound)
			assert.Empty(t, authToken)
		})
	}
}

func TestVerify2FAUserNotActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockOTP := helpers.MockOTP()
	mockUser := helpers.MockUser()
	mockUser.TwoFactorAuth.Enabled = true
	mockUser.IsActive = false
	for _, twoFaTyp := range entity.OtpDeliveryMethods {
		name := "2FA over " + twoFaTyp.String()
		t.Run(name, func(t *testing.T) {
			dto := mockVerify2FAPayload(twoFaTyp)
			authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)
			otpToCompare := dto.Code
			if twoFaTyp == entity.TWO_FA_TOTP_APP {
				otpToCompare = dto.SignInConfToken
			}
			authSuite.mockSecurityProvider.EXPECT().
				ComparePasswords(mockOTP.Code, otpToCompare).
				Return(true, nil)
			authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Email).Return(mockUser, nil)

			authToken, err := authSuite.service.Verify2FA(ctx, dto)
			assert.ErrorIs(t, err, auth.ErrDisabledAccount)
			assert.Empty(t, authToken)
		})
	}
}

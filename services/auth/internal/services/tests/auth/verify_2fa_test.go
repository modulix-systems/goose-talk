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

func mockVerify2FAPayload(twoFaTyp entity.TwoFATransport) *schemas.Verify2FASchema {
	schema := &schemas.Verify2FASchema{
		TwoFATyp: twoFaTyp,
		Code:     "123456",
		Email:    gofakeit.Email(),
		ClientIdentitySchema: schemas.ClientIdentitySchema{
			IPAddr:     gofakeit.IPv4Address(),
			DeviceInfo: gofakeit.UserAgent(),
		},
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
		twoFaTyp      entity.TwoFATransport
		sessionExists bool
	}{
		{
			name:          "2FA over " + string(entity.TWO_FA_EMAIL),
			twoFaTyp:      entity.TWO_FA_EMAIL,
			sessionExists: gofakeit.Bool(),
		},
		{
			name:          "2FA over " + string(entity.TWO_FA_TELEGRAM),
			twoFaTyp:      entity.TWO_FA_TELEGRAM,
			sessionExists: gofakeit.Bool(),
		},
		{
			name:          "2FA over " + string(entity.TWO_FA_TOTP_APP),
			twoFaTyp:      entity.TWO_FA_TOTP_APP,
			sessionExists: gofakeit.Bool(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dto := mockVerify2FAPayload(tc.twoFaTyp)
			mockOTP := helpers.MockOTP()
			mockOTP.UserEmail = dto.Email
			mockUser := helpers.MockUser()
			mockUser.TwoFactorAuth.Enabled = true
			mockSession := helpers.MockUserSession(gofakeit.Bool())
			mockSession.UserId = mockUser.ID
			mockSession.ClientIdentity = &entity.ClientIdentity{DeviceInfo: dto.DeviceInfo, IPAddr: dto.IPAddr}
			authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)
			authSuite.mockCodeRepo.EXPECT().DeleteByEmailOrUserId(ctx, dto.Email, 0).Return(nil)
			otpToCompare := dto.Code
			if tc.twoFaTyp == entity.TWO_FA_TOTP_APP {
				otpToCompare = dto.SignInConfToken
				encryptedTotpSecret := mockUser.TwoFactorAuth.TotpSecret
				plainTotpSecret := string(encryptedTotpSecret)
				authSuite.mockSecurityProvider.EXPECT().DecryptSymmetric(encryptedTotpSecret).Return(plainTotpSecret, nil)
				authSuite.mockSecurityProvider.EXPECT().
					ValidateTOTP(dto.Code, plainTotpSecret).
					Return(true)
			}
			authSuite.mockSecurityProvider.EXPECT().
				ComparePasswords(mockOTP.Code, otpToCompare).
				Return(true, nil)
			setAuthSessionExpectations(t, ctx, authSuite, mockUser, mockSession, tc.sessionExists, true)
			authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Email).Return(mockUser, nil)
			authSuite.mockSecurityProvider.EXPECT().
				GenerateSessionId().Return(mockSession.ID)
			authSession, err := authSuite.service.Verify2FA(ctx, dto)
			assert.NoError(t, err)
			assert.Equal(t, mockSession, authSession)
		})
	}
}

func TestVerify2FAOtpNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	for _, twoFaTyp := range entity.OtpTransports {
		name := "2FA over " + string(twoFaTyp)
		t.Run(name, func(t *testing.T) {
			dto := mockVerify2FAPayload(twoFaTyp)
			authSuite.mockCodeRepo.EXPECT().
				GetByEmail(ctx, dto.Email).
				Return(nil, storage.ErrNotFound)

			authSession, err := authSuite.service.Verify2FA(ctx, dto)
			assert.ErrorIs(t, err, auth.ErrOTPInvalidOrExpired)
			assert.Empty(t, authSession)

		})
	}
}

func TestVerify2FAInvalidOTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockOTP := helpers.MockOTP()
	for _, twoFaTyp := range entity.OtpTransports {
		name := "2FA over " + string(twoFaTyp)
		t.Run(name, func(t *testing.T) {
			dto := mockVerify2FAPayload(twoFaTyp)
			mockOTP.UserEmail = dto.Email
			authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)
			otpToCompare := dto.Code
			if twoFaTyp == entity.TWO_FA_TOTP_APP {
				otpToCompare = dto.SignInConfToken
			}
			authSuite.mockSecurityProvider.EXPECT().
				ComparePasswords(mockOTP.Code, otpToCompare).
				Return(false, nil)

			authSession, err := authSuite.service.Verify2FA(ctx, dto)
			assert.ErrorIs(t, err, auth.ErrOTPInvalidOrExpired)
			assert.Empty(t, authSession)

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
	for _, twoFaTyp := range entity.OtpTransports {
		name := "2FA over " + string(twoFaTyp)
		t.Run(name, func(t *testing.T) {
			dto := mockVerify2FAPayload(twoFaTyp)
			mockOTP.UserEmail = dto.Email
			otpToCompare := dto.Code
			if twoFaTyp == entity.TWO_FA_TOTP_APP {
				otpToCompare = dto.SignInConfToken
			}
			authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)
			authSuite.mockSecurityProvider.EXPECT().
				ComparePasswords(mockOTP.Code, otpToCompare).
				Return(true, nil)
			authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Email).Return(mockUser, nil)

			authSession, err := authSuite.service.Verify2FA(ctx, dto)
			assert.ErrorIs(t, err, auth.Err2FANotEnabled)
			assert.Empty(t, authSession)
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
	mockOTP.UserEmail = dto.Email
	authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)
	authSuite.mockSecurityProvider.EXPECT().
		ComparePasswords(mockOTP.Code, dto.SignInConfToken).
		Return(true, nil)
	authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Email).Return(mockUser, nil)
	encryptedTotpSecret := mockUser.TwoFactorAuth.TotpSecret
	plainTotpSecret := string(encryptedTotpSecret)
	authSuite.mockSecurityProvider.EXPECT().DecryptSymmetric(encryptedTotpSecret).Return(plainTotpSecret, nil)
	authSuite.mockSecurityProvider.EXPECT().
		ValidateTOTP(dto.Code, plainTotpSecret).
		Return(false)

	authSession, err := authSuite.service.Verify2FA(ctx, dto)
	assert.ErrorIs(t, err, auth.ErrOTPInvalidOrExpired)
	assert.Empty(t, authSession)
}

func TestVerify2FAOtpExpired(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockOTP := helpers.MockOTP()
	mockOTP.UpdatedAt = time.Now().Add(-authSuite.mockTTL)
	for _, twoFaTyp := range entity.OtpTransports {
		name := "2FA over " + string(twoFaTyp)
		t.Run(name, func(t *testing.T) {
			dto := mockVerify2FAPayload(twoFaTyp)
			mockOTP.UserEmail = dto.Email
			authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)

			authSession, err := authSuite.service.Verify2FA(ctx, dto)
			assert.ErrorIs(t, err, auth.ErrOTPInvalidOrExpired)
			assert.Empty(t, authSession)
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
	for _, twoFaTyp := range entity.OtpTransports {
		name := "2FA over " + string(twoFaTyp)
		t.Run(name, func(t *testing.T) {
			dto := mockVerify2FAPayload(twoFaTyp)
			mockOTP.UserEmail = dto.Email
			authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)
			otpToCompare := dto.Code
			if twoFaTyp == entity.TWO_FA_TOTP_APP {
				otpToCompare = dto.SignInConfToken
			}
			authSuite.mockSecurityProvider.EXPECT().
				ComparePasswords(mockOTP.Code, otpToCompare).
				Return(true, nil)
			authSuite.mockUsersRepo.EXPECT().GetByLogin(ctx, dto.Email).Return(mockUser, nil)

			authSession, err := authSuite.service.Verify2FA(ctx, dto)
			assert.ErrorIs(t, err, auth.ErrDisabledAccount)
			assert.Empty(t, authSession)
		})
	}
}

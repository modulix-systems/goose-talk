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

func TestConfirm2FASuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockUser := helpers.MockUser()
	mockOTP := helpers.MockOTP()
	confirmationCode := string(mockOTP.Code)
	testCases := []struct {
		twoFaTyp entity.TwoFADeliveryMethod
		contact  string
	}{
		{
			twoFaTyp: entity.TWO_FA_EMAIL,
		},
		{
			twoFaTyp: entity.TWO_FA_EMAIL,
			contact:  gofakeit.Email(),
		},
		{
			twoFaTyp: entity.TWO_FA_TELEGRAM,
		},
		{
			twoFaTyp: entity.TWO_FA_TELEGRAM,
			contact:  gofakeit.Username(),
		},
		{
			twoFaTyp: entity.TWO_FA_TOTP_APP,
		},
		{
			twoFaTyp: entity.TWO_FA_TOTP_APP,
			contact:  gofakeit.Username(),
		},
	}
	for _, tc := range testCases {
		dto := &schemas.Confirm2FASchema{
			UserId:           mockUser.ID,
			Typ:              tc.twoFaTyp,
			ConfirmationCode: confirmationCode,
			Contact:          tc.contact,
		}
		mock2FA := &entity.TwoFactorAuth{
			UserId:         mockUser.ID,
			DeliveryMethod: dto.Typ,
			TotpSecret:     dto.TotpSecret,
			Enabled:        true,
		}
		// accept contact dto field only if typ is email or sms
		if tc.twoFaTyp == entity.TWO_FA_EMAIL || tc.twoFaTyp == entity.TWO_FA_SMS {
			mock2FA.Contact = dto.Contact
		}
		name := "2 fa over " + tc.twoFaTyp.String()
		if tc.contact != "" {
			name += " with contact"
		}
		t.Run(name, func(t *testing.T) {
			if tc.twoFaTyp == entity.TWO_FA_TOTP_APP {
				dto.TotpSecret = gofakeit.UUID()
				authSuite.mockSecurityProvider.EXPECT().ValidateTOTP(dto.ConfirmationCode, dto.TotpSecret).Return(true)
			} else {
				authSuite.mockCodeRepo.EXPECT().GetByUserId(ctx, dto.UserId).Return(mockOTP, nil)
				authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockOTP.Code, dto.ConfirmationCode).Return(true, nil)
			}
			authSuite.mock2FARepo.EXPECT().Insert(ctx, mock2FA).Return(mock2FA, nil)

			res, err := authSuite.service.Confirm2FA(ctx, dto)
			assert.NoError(t, err)
			assert.Equal(t, mockUser.ID, res.UserId)
			assert.Equal(t, tc.twoFaTyp, res.DeliveryMethod)
		})
	}
}

func TestConfirm2FAInvalidOTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	actAndAssert := func(dto *schemas.Confirm2FASchema) {
		res, err := authSuite.service.Confirm2FA(ctx, dto)
		assert.ErrorIs(t, err, auth.ErrOTPInvalidOrExpired)
		assert.Empty(t, res)
	}
	for _, twoFaTyp := range entity.OtpDeliveryMethods {
		mockUser := helpers.MockUser()
		mockOTP := helpers.MockOTP()
		confirmationCode := string(mockOTP.Code)
		dto := &schemas.Confirm2FASchema{
			UserId:           mockUser.ID,
			Typ:              twoFaTyp,
			ConfirmationCode: confirmationCode,
		}
		t.Run("2 fa over "+twoFaTyp.String(), func(t *testing.T) {
			if twoFaTyp == entity.TWO_FA_TOTP_APP {
				t.Run("invalid totp", func(t *testing.T) {
					dto.TotpSecret = gofakeit.UUID()
					authSuite.mockSecurityProvider.EXPECT().ValidateTOTP(dto.ConfirmationCode, dto.TotpSecret).Return(false)
					actAndAssert(dto)
				})
				return
			}
			t.Run("not found otp", func(t *testing.T) {
				authSuite.mockCodeRepo.EXPECT().GetByUserId(ctx, dto.UserId).Return(nil, storage.ErrNotFound)
				actAndAssert(dto)
			})
			t.Run("not matched code", func(t *testing.T) {
				authSuite.mockCodeRepo.EXPECT().GetByUserId(ctx, dto.UserId).Return(mockOTP, nil)
				authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockOTP.Code, dto.ConfirmationCode).Return(false, nil)
				actAndAssert(dto)
			})
			t.Run("expired otp", func(t *testing.T) {
				mockOTP.UpdatedAt = time.Now().Add(-authSuite.tokenTTL)
				authSuite.mockCodeRepo.EXPECT().GetByUserId(ctx, dto.UserId).Return(mockOTP, nil)
				actAndAssert(dto)
			})
		})
	}
}

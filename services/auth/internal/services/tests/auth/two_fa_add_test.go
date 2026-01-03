package auth_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/schemas"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAdd2FASuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockUser := helpers.MockUser()
	plainOTPCode := gofakeit.Numerify("######")
	setOTPExpectations := func(otp *entity.OTP) {
		authSuite.mockSecurityProvider.EXPECT().GenerateOTPCode().Return(plainOTPCode)
		authSuite.mockSecurityProvider.EXPECT().HashPassword(plainOTPCode).Return(otp.Code, nil)
		authSuite.mockCodeRepo.EXPECT().InsertOrUpdateCode(ctx, otp).Return(nil)
	}
	mockUser.TwoFactorAuth.Enabled = false
	t.Run("Add 2fa by email", func(t *testing.T) {
		for _, contact := range []string{gofakeit.Email(), ""} {
			t.Run("With contact: "+contact, func(t *testing.T) {
				email := contact
				if email == "" {
					email = mockUser.Email
				}
				mockOTP := &entity.OTP{UserId: mockUser.Id, Code: []byte(plainOTPCode)}
				authSuite.mockMailSender.EXPECT().Send2FAEmail(ctx, email, plainOTPCode)
				authSuite.mockUsersRepo.EXPECT().GetByID(ctx, mockUser.Id).Return(mockUser, nil)
				setOTPExpectations(mockOTP)

				connInfo, err := authSuite.service.Add2FA(ctx, &schemas.Add2FADto{UserId: mockUser.Id, Typ: entity.TWO_FA_EMAIL, Contact: contact})

				assert.Empty(t, connInfo)
				assert.NoError(t, err)

			})
		}
	})
	t.Run("Add 2fa by tg", func(t *testing.T) {
		linkCode := gofakeit.Numerify("######")
		expectedLink := "https://example.com?" + linkCode
		mockTgMsg := &gateways.TelegramMsg{
			ChatId:   gofakeit.UUID(),
			Text:     "/start " + linkCode,
			DateSent: time.Now().Add(time.Second),
		}
		mockOTP := &entity.OTP{UserId: mockUser.Id, Code: []byte(plainOTPCode)}
		authSuite.mockUsersRepo.EXPECT().GetByID(ctx, mockUser.Id).Return(mockUser, nil)
		setOTPExpectations(mockOTP)
		authSuite.mockSecurityProvider.EXPECT().GenerateOTPCode().Return(linkCode)
		authSuite.mockTgAPI.EXPECT().GetStartLinkWithCode(linkCode).Return(expectedLink)
		authSuite.mockTgAPI.EXPECT().GetLatestMsg(ctx).Return(mockTgMsg, nil)
		authSuite.mock2FARepo.EXPECT().UpdateContactForUser(ctx, mockUser.Id, mockTgMsg.ChatId).Return(nil)
		authSuite.mockTgAPI.EXPECT().SendTextMsg(ctx, mockTgMsg.ChatId, fmt.Sprintf("Authorization code: %s", plainOTPCode))

		connInfo, err := authSuite.service.Add2FA(ctx, &schemas.Add2FADto{UserId: mockUser.Id, Typ: entity.TWO_FA_TELEGRAM})
		// wait some time for goroutine to complete
		time.Sleep(500 * time.Millisecond)

		require.NotNil(t, connInfo)
		assert.Equal(t, expectedLink, connInfo.Link)
		assert.Empty(t, connInfo.TotpSecret)
		assert.NoError(t, err)
	})

	t.Run("Add 2fa by totp", func(t *testing.T) {
		expectedSecret := gofakeit.UUID()
		expectedLink := gofakeit.URL()
		authSuite.mockUsersRepo.EXPECT().GetByID(ctx, mockUser.Id).Return(mockUser, nil)
		authSuite.mockSecurityProvider.EXPECT().GenerateTOTPEnrollUrlWithSecret(mockUser.Email).Return(expectedLink, expectedSecret)

		connInfo, err := authSuite.service.Add2FA(ctx, &schemas.Add2FADto{UserId: mockUser.Id, Typ: entity.TWO_FA_TOTP_APP})

		require.NotNil(t, connInfo)
		assert.Equal(t, expectedLink, connInfo.Link)
		assert.Equal(t, expectedSecret, connInfo.TotpSecret)
		assert.NoError(t, err)
	})
}

func TestAdd2FAUnsupportedTyp(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockUser := helpers.MockUser()
	mockUser.TwoFactorAuth = nil
	const unsupported2FAMethod entity.TwoFATransport = "unsupported"
	authSuite.mockUsersRepo.EXPECT().GetByID(ctx, mockUser.Id).Return(mockUser, nil)

	connInfo, err := authSuite.service.Add2FA(ctx, &schemas.Add2FADto{UserId: mockUser.Id, Typ: unsupported2FAMethod})

	assert.Empty(t, connInfo)
	assert.ErrorIs(t, err, auth.ErrUnsupported2FAMethod)
}

func TestAdd2FaAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockUser := helpers.MockUser()
	mockUser.TwoFactorAuth.Enabled = true
	authSuite.mockUsersRepo.EXPECT().GetByID(ctx, mockUser.Id).Return(mockUser, nil)

	connInfo, err := authSuite.service.Add2FA(ctx, &schemas.Add2FADto{UserId: mockUser.Id, Typ: helpers.RandomChoose(entity.OtpTransports...)})

	assert.Empty(t, connInfo)
	assert.ErrorIs(t, err, auth.Err2FaAlreadyAdded)
}

func TestAdd2FaUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockUser := helpers.MockUser()
	authSuite.mockUsersRepo.EXPECT().GetByID(ctx, mockUser.Id).Return(nil, storage.ErrNotFound)

	connInfo, err := authSuite.service.Add2FA(ctx, &schemas.Add2FADto{UserId: mockUser.Id, Typ: helpers.RandomChoose(entity.OtpTransports...)})

	assert.Empty(t, connInfo)
	assert.ErrorIs(t, err, auth.ErrUserNotFound)
}

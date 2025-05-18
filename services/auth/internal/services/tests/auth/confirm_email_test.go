package auth_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestConfirmEmailSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	mockEmail := gofakeit.Email()
	plainOTPCode := "securetoken"
	hashedOTPCode := []byte(plainOTPCode)
	mockOTP := &entity.OTP{Code: hashedOTPCode, UserEmail: mockEmail}
	ctx := context.Background()
	authSuite.mockUsersRepo.EXPECT().CheckExistsWithEmail(ctx, mockEmail).Return(false, nil)
	authSuite.mockCodeRepo.EXPECT().InsertOrUpdateCode(ctx, mockOTP).Return(nil)
	authSuite.mockMailSender.EXPECT().SendSignUpConfirmationEmail(ctx, mockEmail, plainOTPCode).Return(nil)
	authSuite.mockSecurityProvider.EXPECT().GenerateOTPCode().Return(plainOTPCode)
	authSuite.mockSecurityProvider.EXPECT().HashPassword(plainOTPCode).Return(hashedOTPCode, nil)

	err := authSuite.service.ConfirmEmail(ctx, mockEmail)

	assert.NoError(t, err)
}

func TestConfirmEmailUserExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	mockEmail := gofakeit.Email()
	ctx := context.Background()
	authSuite.mockUsersRepo.EXPECT().CheckExistsWithEmail(ctx, mockEmail).Return(true, nil)

	err := authSuite.service.ConfirmEmail(ctx, mockEmail)

	assert.ErrorIs(t, err, auth.ErrUserAlreadyExists)
}

func TestConfirmEmailCodeAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	mockEmail := gofakeit.Email()
	otpCode := "testcode"
	mockOTP := &entity.OTP{Code: []byte(otpCode), UserEmail: mockEmail}
	ctx := context.Background()
	authSuite.mockSecurityProvider.EXPECT().HashPassword(otpCode).Return(mockOTP.Code, nil)
	authSuite.mockUsersRepo.EXPECT().CheckExistsWithEmail(ctx, mockEmail).Return(false, nil)
	authSuite.mockCodeRepo.EXPECT().InsertOrUpdateCode(ctx, mockOTP).Return(nil)
	authSuite.mockMailSender.EXPECT().SendSignUpConfirmationEmail(ctx, mockEmail, otpCode).Return(nil)
	authSuite.mockSecurityProvider.EXPECT().GenerateOTPCode().Return(otpCode)

	err := authSuite.service.ConfirmEmail(ctx, mockEmail)

	assert.NoError(t, err)

}

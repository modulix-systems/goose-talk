package auth_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestConfirmEmailSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	mockEmail := gofakeit.Email()
	mockCode := "securetoken"
	mockSignUpCode := &entity.SignUpCode{Code: mockCode, Email: mockEmail}
	ctx := context.Background()
	authSuite.mockUsersRepo.EXPECT().CheckExistsWithEmail(ctx, mockEmail).Return(false, nil)
	authSuite.mockCodeRepo.EXPECT().Insert(ctx, mockSignUpCode).Return(nil)
	authSuite.mockMailSender.EXPECT().SendSignUpConfirmationEmail(ctx, mockEmail, mockSignUpCode.Code).Return(nil)
	authSuite.mockSecurityProvider.EXPECT().NewSecureToken(6).Return(mockCode)

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
	mockSignUpCode := &entity.SignUpCode{Code: "secureToken", Email: mockEmail}
	mockSignUpCodeFromDB := &entity.SignUpCode{Code: "anothersecuretoken", Email: mockEmail}
	ctx := context.Background()
	authSuite.mockUsersRepo.EXPECT().CheckExistsWithEmail(ctx, mockEmail).Return(false, nil)
	authSuite.mockCodeRepo.EXPECT().Insert(ctx, mockSignUpCode).Return(storage.ErrAlreadyExists)
	authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, mockEmail).Return(mockSignUpCodeFromDB, nil)
	authSuite.mockMailSender.EXPECT().SendSignUpConfirmationEmail(ctx, mockEmail, mockSignUpCodeFromDB.Code).Return(nil)
	authSuite.mockSecurityProvider.EXPECT().NewSecureToken(6).Return(mockSignUpCode.Code)

	err := authSuite.service.ConfirmEmail(ctx, mockEmail)

	assert.NoError(t, err)

}

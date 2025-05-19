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

func mockSignUpPayload() *schemas.SignUpSchema {
	return &schemas.SignUpSchema{
		Username:         gofakeit.Username(),
		Email:            gofakeit.Email(),
		FirstName:        helpers.RandomChoose(gofakeit.FirstName(), ""),
		LastName:         helpers.RandomChoose(gofakeit.LastName(), ""),
		ConfirmationCode: gofakeit.Numerify("######"),
		Password:         helpers.RandomPassword(),
	}
}

func TestSignupSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	dto := mockSignUpPayload()
	expectedToken := "auth token"
	mockOTP := helpers.MockOTP()
	mockOTP.UserEmail = dto.Email
	ctx := context.Background()
	userToInsert := &entity.User{FirstName: dto.FirstName, LastName: dto.LastName, Email: dto.Email, Password: []byte(dto.Password)}
	insertedUser := *userToInsert
	insertedUser.ID = gofakeit.Number(1, 1000)
	authSuite.mockAuthTokenProvider.EXPECT().
		NewToken(authSuite.tokenTTL, map[string]any{"uid": insertedUser.ID}).
		Return(expectedToken, nil)
	authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockOTP.Code, dto.ConfirmationCode).Return(true, nil)
	authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)
	authSuite.mockSecurityProvider.EXPECT().HashPassword(dto.Password).Return(userToInsert.Password, nil)
	authSuite.mockUsersRepo.EXPECT().
		Insert(ctx, userToInsert).
		Return(&insertedUser, nil)
	expectedName := dto.Username
	if dto.FirstName != "" {
		expectedName = dto.FirstName
		if dto.LastName != "" {
			expectedName = expectedName + " " + dto.LastName
		}
	}
	authSuite.mockMailSender.EXPECT().SendGreetingEmail(ctx, dto.Email, expectedName)

	token, user, err := authSuite.service.SignUp(ctx, dto)

	assert.Equal(t, token, expectedToken)
	assert.Equal(t, user.ID, insertedUser.ID)
	assert.NoError(t, err)
}

func TestSignupNotFoundCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	dto := mockSignUpPayload()
	ctx := context.Background()
	authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(
		nil,
		storage.ErrNotFound,
	)

	token, user, err := authSuite.service.SignUp(ctx, dto)

	assert.Empty(t, token)
	assert.Empty(t, user)
	assert.ErrorIs(t, err, auth.ErrInvalidOtp)
}

func TestSignupExpiredCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	dto := mockSignUpPayload()
	ctx := context.Background()
	hashedCode := []byte("hashedCode")
	authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(
		&entity.OTP{
			Code:      hashedCode,
			UserEmail: dto.Email,
			CreatedAt: time.Now().Add(-authSuite.tokenTTL / 2),
			UpdatedAt: time.Now().Add(-authSuite.tokenTTL),
		},
		nil,
	)

	token, user, err := authSuite.service.SignUp(ctx, dto)

	assert.Empty(t, token)
	assert.Empty(t, user)
	assert.ErrorIs(t, err, auth.ErrOtpExpired)
}

func TestSignUpUserExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	dto := mockSignUpPayload()
	ctx := context.Background()
	mockOTP := helpers.MockOTP()
	mockOTP.UserEmail = dto.Email
	authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)
	hashedPassword := []byte("hashedPassword")
	authSuite.mockSecurityProvider.EXPECT().ComparePasswords(mockOTP.Code, dto.ConfirmationCode).Return(true, nil)
	authSuite.mockSecurityProvider.EXPECT().HashPassword(dto.Password).Return(hashedPassword, nil)
	authSuite.mockUsersRepo.EXPECT().
		Insert(ctx, &entity.User{FirstName: dto.FirstName, LastName: dto.LastName, Email: dto.Email, Password: hashedPassword}).
		Return(nil, storage.ErrAlreadyExists)

	token, user, err := authSuite.service.SignUp(ctx, dto)

	assert.Empty(t, token)
	assert.Empty(t, user)
	assert.ErrorIs(t, err, auth.ErrUserAlreadyExists)

}

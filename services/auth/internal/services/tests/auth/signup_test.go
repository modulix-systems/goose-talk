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
		LoginInfoSchema: schemas.LoginInfoSchema{
			DeviceInfo: gofakeit.UserAgent(),
			IPAddr:     gofakeit.IPv4Address(),
		},
	}
}

func TestSignupSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	dto := mockSignUpPayload()
	mockOTP := helpers.MockOTP()
	mockOTP.UserEmail = dto.Email
	ctx := context.Background()
	userToInsert := &entity.User{
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		Email:     dto.Email,
		Password:  []byte(dto.Password),
	}
	insertedUser := *userToInsert
	insertedUser.ID = gofakeit.Number(1, 1000)
	expectedSession := helpers.MockAuthSession(true)
	expectedSession.UserId = insertedUser.ID
	setExpectations := func(rememberMe bool) {
		dto.RememberMe = rememberMe
		authSuite.mockCodeRepo.EXPECT().GetByEmail(ctx, dto.Email).Return(mockOTP, nil)
		authSuite.mockCodeRepo.EXPECT().DeleteByEmailOrUserId(ctx, dto.Email, 0).Return(nil)
		authSuite.mockSecurityProvider.EXPECT().
			ComparePasswords(mockOTP.Code, dto.ConfirmationCode).
			Return(true, nil)
		authSuite.mockSecurityProvider.EXPECT().
			GenerateSessionId().
			Return(expectedSession.ID)
		authSuite.mockSecurityProvider.EXPECT().
			HashPassword(dto.Password).
			Return(userToInsert.Password, nil)
		authSuite.mockUsersRepo.EXPECT().
			Insert(ctx, userToInsert).
			Return(&insertedUser, nil)
		expectedLocation := gofakeit.City()
		authSuite.mockGeoIPApi.EXPECT().GetLocationByIP(dto.IPAddr).Return(expectedLocation, nil)
		authSuite.mockSessionsRepo.EXPECT().Insert(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, session *entity.AuthSession) (*entity.AuthSession, error) {
				assert.Equal(t, expectedSession.UserId, session.UserId)
				assert.Equal(t, expectedSession.ID, session.ID)
				assert.Equal(t, dto.IPAddr, session.ClientIdentity.IPAddr)
				assert.Equal(t, dto.DeviceInfo, session.ClientIdentity.DeviceInfo)
				assert.Equal(t, expectedLocation, session.ClientIdentity.Location)
				sessionTTL := authSuite.mockTTL
				if rememberMe {
					sessionTTL = authSuite.longLivedSessionTTL
				}
				assert.WithinDuration(t, time.Now().Add(sessionTTL), session.ExpiresAt, time.Second)

				return expectedSession, nil
			})
		expectedName := dto.Username
		if dto.FirstName != "" {
			expectedName = dto.FirstName
			if dto.LastName != "" {
				expectedName = expectedName + " " + dto.LastName
			}
		}
		authSuite.mockMailSender.EXPECT().SendGreetingEmail(ctx, dto.Email, expectedName)
	}
	makeAssertions := func(authSession *entity.AuthSession, err error) {
		assert.NoError(t, err)
		assert.True(t, authSession.IsActive())
		assert.Equal(t, expectedSession.ID, authSession.ID)
		assert.Equal(t, insertedUser, *authSession.User)
	}
	t.Run("Short lived session", func(t *testing.T) {
		setExpectations(false)
		authSession, err := authSuite.service.SignUp(ctx, dto)
		makeAssertions(authSession, err)
	})
	t.Run("Long lived session", func(t *testing.T) {
		setExpectations(true)
		authSession, err := authSuite.service.SignUp(ctx, dto)
		makeAssertions(authSession, err)
	})
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

	authSession, err := authSuite.service.SignUp(ctx, dto)

	assert.Empty(t, authSession)
	assert.ErrorIs(t, err, auth.ErrOtpIsNotValid)
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
			CreatedAt: time.Now().Add(-authSuite.mockTTL / 2),
			UpdatedAt: time.Now().Add(-authSuite.mockTTL),
		},
		nil,
	)

	authSession, err := authSuite.service.SignUp(ctx, dto)

	assert.Empty(t, authSession)
	assert.ErrorIs(t, err, auth.ErrOtpIsNotValid)
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
	authSuite.mockSecurityProvider.EXPECT().
		ComparePasswords(mockOTP.Code, dto.ConfirmationCode).
		Return(true, nil)
	authSuite.mockSecurityProvider.EXPECT().HashPassword(dto.Password).Return(hashedPassword, nil)
	authSuite.mockUsersRepo.EXPECT().
		Insert(ctx, &entity.User{FirstName: dto.FirstName, LastName: dto.LastName, Email: dto.Email, Password: hashedPassword}).
		Return(nil, storage.ErrAlreadyExists)

	authSession, err := authSuite.service.SignUp(ctx, dto)

	assert.Empty(t, authSession)
	assert.ErrorIs(t, err, auth.ErrUserAlreadyExists)
}

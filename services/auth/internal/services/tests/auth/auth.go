package auth_test

import (
	"time"

	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/tests/mocks"
	"github.com/modulix-systems/goose-talk/tests/suite"
	"go.uber.org/mock/gomock"
)

type AuthTestSuite struct {
	mockCodeRepo          *mocks.MockSignUpCodeRepo
	mockUsersRepo         *mocks.MockUsersRepo
	mockAuthTokenProvider *mocks.MockAuthTokenProvider
	authService           *auth.AuthService
	tokenTTL              time.Duration
}

func NewAuthTestSuite(ctrl *gomock.Controller) *AuthTestSuite {
	mockCodeRepo := mocks.NewMockSignUpCodeRepo(ctrl)
	mockUsersRepo := mocks.NewMockUsersRepo(ctrl)
	tokenTTL := suite.MockDuration("")
	mockAuthTokenProvider := mocks.NewMockAuthTokenProvider(ctrl)
	service := auth.New(
		mockUsersRepo,
		mockCodeRepo,
		mockAuthTokenProvider,
		tokenTTL,
		tokenTTL,
	)
	return &AuthTestSuite{
		mockCodeRepo:          mockCodeRepo,
		mockUsersRepo:         mockUsersRepo,
		mockAuthTokenProvider: mockAuthTokenProvider,
		tokenTTL:              tokenTTL,
		authService:           service,
	}
}

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
	mockMailSender        *mocks.MockNotificationsService
	mockSecurityProvider  *mocks.MockSecurityProvider
	service               *auth.AuthService
	tokenTTL              time.Duration
}

func NewAuthTestSuite(ctrl *gomock.Controller) *AuthTestSuite {
	mockCodeRepo := mocks.NewMockSignUpCodeRepo(ctrl)
	mockUsersRepo := mocks.NewMockUsersRepo(ctrl)
	tokenTTL := suite.MockDuration("")
	mockAuthTokenProvider := mocks.NewMockAuthTokenProvider(ctrl)
	mockMailSender := mocks.NewMockNotificationsService(ctrl)
	mockSecurityProvider := mocks.NewMockSecurityProvider(ctrl)
	service := auth.New(
		mockUsersRepo,
		mockMailSender,
		mockCodeRepo,
		mockAuthTokenProvider,
		tokenTTL,
		tokenTTL,
		mockSecurityProvider,
	)
	return &AuthTestSuite{
		mockCodeRepo:          mockCodeRepo,
		mockUsersRepo:         mockUsersRepo,
		mockSecurityProvider:  mockSecurityProvider,
		mockAuthTokenProvider: mockAuthTokenProvider,
		tokenTTL:              tokenTTL,
		mockMailSender:        mockMailSender,
		service:               service,
	}
}

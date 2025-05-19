package auth_test

import (
	"time"

	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/tests/mocks"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"go.uber.org/mock/gomock"
)

type AuthTestSuite struct {
	mockCodeRepo          *mocks.MockOtpRepo
	mockUsersRepo         *mocks.MockUsersRepo
	mockSessionsRepo      *mocks.MockUserSessionsRepo
	mockAuthTokenProvider *mocks.MockAuthTokenProvider
	mockMailSender        *mocks.MockNotificationsService
	mockSecurityProvider  *mocks.MockSecurityProvider
	mockTgAPI             *mocks.MockTelegramBotAPI
	mockGeoIPApi          *mocks.MockGeoIPApi
	service               *auth.AuthService
	tokenTTL              time.Duration
}

func NewAuthTestSuite(ctrl *gomock.Controller) *AuthTestSuite {
	mockCodeRepo := mocks.NewMockOtpRepo(ctrl)
	mockUsersRepo := mocks.NewMockUsersRepo(ctrl)
	tokenTTL := helpers.MockDuration("")
	mockAuthTokenProvider := mocks.NewMockAuthTokenProvider(ctrl)
	mockMailSender := mocks.NewMockNotificationsService(ctrl)
	mockSecurityProvider := mocks.NewMockSecurityProvider(ctrl)
	mockTgAPI := mocks.NewMockTelegramBotAPI(ctrl)
	mockSessionsRepo := mocks.NewMockUserSessionsRepo(ctrl)
	mockGeoIPApi := mocks.NewMockGeoIPApi(ctrl)
	service := auth.New(
		mockUsersRepo,
		mockMailSender,
		mockCodeRepo,
		mockAuthTokenProvider,
		tokenTTL,
		tokenTTL,
		mockSecurityProvider,
		mockTgAPI,
		mockSessionsRepo,
		mockGeoIPApi,
	)
	return &AuthTestSuite{
		mockCodeRepo:          mockCodeRepo,
		mockUsersRepo:         mockUsersRepo,
		mockSecurityProvider:  mockSecurityProvider,
		mockAuthTokenProvider: mockAuthTokenProvider,
		mockSessionsRepo:      mockSessionsRepo,
		tokenTTL:              tokenTTL,
		mockMailSender:        mockMailSender,
		mockTgAPI:             mockTgAPI,
		service:               service,
		mockGeoIPApi:          mockGeoIPApi,
	}
}

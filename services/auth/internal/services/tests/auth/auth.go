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
	"github.com/modulix-systems/goose-talk/tests/mocks"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type AuthTestSuite struct {
	mockCodeRepo          *mocks.MockOtpRepo
	mock2FARepo           *mocks.MockTwoFactorAuthRepo
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
	mock2FARepo := mocks.NewMockTwoFactorAuthRepo(ctrl)
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
		mock2FARepo,
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
		mock2FARepo:           mock2FARepo,
	}
}

func setAuthSessionExpectations(t *testing.T, ctx context.Context, authSuite *AuthTestSuite, mockUser *entity.User, mockSession *entity.UserSession, sessionExists bool, deviceInfo string, ip string, authToken string) {
	t.Helper()
	mockLocation := gofakeit.City()
	authSuite.mockGeoIPApi.EXPECT().GetLocationByIP(ip).Return(mockLocation, nil)
	if sessionExists {
		authSuite.mockSessionsRepo.EXPECT().
			GetByParamsMatch(ctx, ip, deviceInfo, mockUser.ID).
			Return(mockSession, nil)
		authSuite.mockSessionsRepo.EXPECT().UpdateById(
			ctx, mockSession.ID,
			gomock.Any()).
			DoAndReturn(func(ctx context.Context, sessionId int, payload *schemas.SessionUpdatePayload) (*entity.UserSession, error) {
				require.NotNil(t, payload)
				assert.NotNil(t, payload.DeactivatedAt)
				assert.Equal(t, *payload.DeactivatedAt, time.Time{})
				assert.WithinDuration(t, time.Now(), payload.LastSeenAt, time.Second)
				assert.Equal(t, authToken, payload.AccessToken)
				return mockSession, nil
			})
	} else {
		authSuite.mockSessionsRepo.EXPECT().GetByParamsMatch(ctx, ip, deviceInfo, mockUser.ID).Return(nil, storage.ErrNotFound)
		authSuite.mockSessionsRepo.EXPECT().Insert(ctx, &entity.UserSession{
			UserId:      mockUser.ID,
			DeviceInfo:  deviceInfo,
			IP:          ip,
			Location:    mockLocation,
			AccessToken: authToken,
		}).Return(mockSession, nil)
		authSuite.mockMailSender.EXPECT().SendSignInNewDeviceEmail(ctx, mockUser.Email, mockSession)
	}

}

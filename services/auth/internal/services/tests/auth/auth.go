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
	mockLoginTokenRepo    *mocks.MockLoginTokenRepo
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
	mockLoginTokenRepo := mocks.NewMockLoginTokenRepo(ctrl)
	mock2FARepo := mocks.NewMockTwoFactorAuthRepo(ctrl)
	mockGeoIPApi := mocks.NewMockGeoIPApi(ctrl)
	service := auth.New(
		mockUsersRepo,
		mockMailSender,
		mockCodeRepo,
		mockAuthTokenProvider,
		tokenTTL,
		tokenTTL,
		tokenTTL,
		mockSecurityProvider,
		mockTgAPI,
		mockSessionsRepo,
		mockGeoIPApi,
		mock2FARepo,
		mockLoginTokenRepo,
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
		mockLoginTokenRepo:    mockLoginTokenRepo,
	}
}

func setAuthSessionExpectations(t *testing.T, ctx context.Context, authSuite *AuthTestSuite, mockUser *entity.User, mockSession *entity.UserSession, sessionExists bool, shouldFetchLocation bool) {
	t.Helper()
	expectedIP := mockSession.ClientIdentity.IPAddr
	expectedDeviceInfo := mockSession.ClientIdentity.DeviceInfo
	expectedLocation := mockSession.ClientIdentity.Location
	if shouldFetchLocation {
		expectedLocation = gofakeit.City()
		authSuite.mockGeoIPApi.EXPECT().GetLocationByIP(expectedIP).Return(expectedLocation, nil)
		mockSession.ClientIdentity.Location = expectedLocation
	}
	if sessionExists {
		authSuite.mockSessionsRepo.EXPECT().
			GetByParamsMatch(ctx, expectedIP, expectedDeviceInfo, mockUser.ID).
			Return(mockSession, nil)
		authSuite.mockSessionsRepo.EXPECT().UpdateById(
			ctx, mockSession.ID,
			gomock.Any()).
			DoAndReturn(func(ctx context.Context, sessionId int, payload *schemas.SessionUpdatePayload) (*entity.UserSession, error) {
				require.NotNil(t, payload)
				assert.NotNil(t, payload.DeactivatedAt)
				assert.Equal(t, *payload.DeactivatedAt, time.Time{})
				assert.WithinDuration(t, time.Now(), payload.LastSeenAt, time.Second)
				assert.Equal(t, mockSession.AccessToken, payload.AccessToken)
				return mockSession, nil
			})
	} else {
		authSuite.mockSessionsRepo.EXPECT().GetByParamsMatch(ctx, expectedIP, expectedDeviceInfo, mockUser.ID).Return(nil, storage.ErrNotFound)
		authSuite.mockSessionsRepo.EXPECT().Insert(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, session *entity.UserSession) (*entity.UserSession, error) {
				assert.Equal(t, mockSession.UserId, session.UserId)
				assert.Equal(t, mockSession.AccessToken, session.AccessToken)
				// if related entity id is not provided - assert that has enought data to create that related entity
				if session.ClientIdentityId == 0 {
					assert.Equal(t, expectedIP, session.ClientIdentity.IPAddr)
					assert.Equal(t, expectedDeviceInfo, session.ClientIdentity.DeviceInfo)
					assert.Equal(t, expectedLocation, session.ClientIdentity.Location)
				}
				return mockSession, nil
			})
		authSuite.mockMailSender.EXPECT().SendSignInNewDeviceEmail(ctx, mockUser.Email, mockSession)
	}

}

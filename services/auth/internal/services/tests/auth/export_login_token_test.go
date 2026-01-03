package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/schemas"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func fakeExportLoginTokenPayload(fromEnt *entity.QRCodeLoginToken) *schemas.ExportLoginTokenDto {
	return &schemas.ExportLoginTokenDto{
		ClientId: fromEnt.ClientId,
		LoginInfoSchema: schemas.LoginInfoSchema{
			IPAddr:     fromEnt.ClientIdentity.IPAddr,
			DeviceInfo: fromEnt.ClientIdentity.DeviceInfo,
		},
	}
}

func setInsertExpectation(t *testing.T, ctx context.Context, authSuite *AuthTestSuite, expectedLoginToken *entity.QRCodeLoginToken, withClientId bool) {
	t.Helper()
	tokenVal := gofakeit.LetterN(16)
	authSuite.mockSecurityProvider.EXPECT().GenerateSecretTokenUrlSafe(16).Return(tokenVal)
	authSuite.mockLoginTokenRepo.EXPECT().Insert(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, token *entity.QRCodeLoginToken) (*entity.QRCodeLoginToken, error) {
			assert.Equal(t, expectedLoginToken.ClientId, token.ClientId)
			assert.Equal(t, tokenVal, token.Value)
			assert.WithinDuration(t, time.Now().Add(authSuite.mockTTL), token.ExpiresAt, time.Second)
			if withClientId {
				assert.Equal(t, expectedLoginToken.ClientIdentity.IPAddr, token.ClientIdentity.IPAddr)
				assert.Equal(t, expectedLoginToken.ClientIdentity.Location, token.ClientIdentity.Location)
			} else {
				assert.Empty(t, token.ClientIdentity)
				assert.Equal(t, expectedLoginToken.ClientIdentityId, token.ClientIdentityId)
			}
			return expectedLoginToken, nil
		})

}

func TestExportLoginTokenInsertNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockLoginToken := helpers.MockLoginToken(authSuite.mockTTL)
	mockLoginToken.AuthSessionId = 0
	dto := fakeExportLoginTokenPayload(mockLoginToken)

	authSuite.mockLoginTokenRepo.EXPECT().GetByClientId(ctx, dto.ClientId).Return(nil, storage.ErrNotFound)
	authSuite.mockGeoIPApi.EXPECT().GetLocationByIP(dto.IPAddr).Return(mockLoginToken.ClientIdentity.Location, nil)
	setInsertExpectation(t, ctx, authSuite, mockLoginToken, true)

	token, err := authSuite.service.ExportLoginToken(ctx, dto)
	assert.NoError(t, err)
	assert.False(t, token.IsApproved())
}

func TestExportLoginTokenReInsertNotApproved(t *testing.T) {
	// If token for session already exists but not approved yet - delete it and create new one
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockLoginToken := helpers.MockLoginToken(authSuite.mockTTL)
	mockLoginToken.AuthSessionId = 0
	dto := fakeExportLoginTokenPayload(mockLoginToken)
	authSuite.mockLoginTokenRepo.EXPECT().GetByClientId(ctx, dto.ClientId).Return(mockLoginToken, nil)
	authSuite.mockLoginTokenRepo.EXPECT().DeleteByClientId(ctx, dto.ClientId).Return(nil)
	setInsertExpectation(t, ctx, authSuite, mockLoginToken, false)

	token, err := authSuite.service.ExportLoginToken(ctx, dto)
	assert.NoError(t, err)
	assert.Equal(t, mockLoginToken.ClientId, token.ClientId)
	assert.False(t, token.IsApproved())
}

func TestExportLoginTokenReturnApproved(t *testing.T) {
	// If token for session already approved - return it
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockLoginToken := helpers.MockLoginToken(authSuite.mockTTL)
	dto := fakeExportLoginTokenPayload(mockLoginToken)
	authSuite.mockLoginTokenRepo.EXPECT().GetByClientId(ctx, dto.ClientId).Return(mockLoginToken, nil)

	token, err := authSuite.service.ExportLoginToken(ctx, dto)
	assert.NoError(t, err)
	assert.Equal(t, mockLoginToken.ClientId, token.ClientId)
	assert.NotEmpty(t, token.AuthSession)
	assert.True(t, token.IsApproved())
}

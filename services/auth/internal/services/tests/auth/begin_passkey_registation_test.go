package auth_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/gateways"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestBeginPasskeyRegistrationSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	fakeUser := helpers.MockUser()
	expectedOptions := gateways.WebAuthnRegistrationOptions(gofakeit.Sentence(10))
	fakePasskeySession := helpers.MockPasskeySession()
	serializedPasskeySession, err := json.Marshal(fakePasskeySession)
	require.NoError(t, err)
	authSuite.mockUsersRepo.EXPECT().GetByIDWithPasskeyCredentials(ctx, fakeUser.Id).Return(fakeUser, nil)
	authSuite.mockWebAuthnProvider.EXPECT().GenerateRegistrationOptions(fakeUser).Return(expectedOptions, fakePasskeySession, nil)
	authSuite.mockKeyValueStorage.EXPECT().Set(fmt.Sprintf("passkey_session:%d", fakeUser.Id), string(serializedPasskeySession), time.Duration(0)).Return(nil)

	options, err := authSuite.service.BeginPasskeyRegistration(ctx, fakeUser.Id)

	assert.Equal(t, expectedOptions, options)
	assert.NoError(t, err)
}

func TestBeginPasskeyRegistrationUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	fakeUser := helpers.MockUser()
	authSuite.mockUsersRepo.EXPECT().GetByIDWithPasskeyCredentials(ctx, fakeUser.Id).Return(nil, storage.ErrNotFound)

	options, err := authSuite.service.BeginPasskeyRegistration(ctx, fakeUser.Id)

	assert.Nil(t, options)
	assert.ErrorIs(t, err, auth.ErrUserNotFound)
}

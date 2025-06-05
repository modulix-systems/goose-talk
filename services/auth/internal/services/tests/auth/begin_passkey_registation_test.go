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

func createFakePasskeySession() *gateways.PasskeyTmpSession {
	return &gateways.PasskeyTmpSession{
		UserId:    []byte(gofakeit.Numerify("###")),
		Challenge: gofakeit.Sentence(10),
		CredParams: []gateways.PasskeyCredentialParam{
			gateways.PasskeyCredentialParam{Type: gofakeit.AppName(), Alg: gofakeit.Number(1, 10)},
			gateways.PasskeyCredentialParam{Type: gofakeit.AppName(), Alg: gofakeit.Number(1, 10)},
		},
	}
}

func TestBeginPasskeyRegistrationSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	fakeUser := helpers.MockUser()
	expectedOptions := gateways.WebAuthnRegistrationOptions(gofakeit.Sentence(10))
	fakePasskeySession := createFakePasskeySession()
	serializedPasskeySession, err := json.Marshal(fakePasskeySession)
	require.NoError(t, err)
	authSuite.mockUsersRepo.EXPECT().GetByID(ctx, fakeUser.ID).Return(fakeUser, nil)
	authSuite.mockWebAuthnProvider.EXPECT().GenerateRegistrationOptions(fakeUser).Return(expectedOptions, fakePasskeySession, nil)
	authSuite.mockKeyValueStorage.EXPECT().Set(fmt.Sprintf("passkey_session:%d", fakeUser.ID), string(serializedPasskeySession), time.Duration(0)).Return(nil)

	options, err := authSuite.service.BeginPasskeyRegistration(ctx, fakeUser.ID)

	assert.Equal(t, expectedOptions, options)
	assert.NoError(t, err)
}

func TestBeginPasskeyRegistrationUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	fakeUser := helpers.MockUser()
	authSuite.mockUsersRepo.EXPECT().GetByID(ctx, fakeUser.ID).Return(nil, storage.ErrNotFound)

	options, err := authSuite.service.BeginPasskeyRegistration(ctx, fakeUser.ID)

	assert.Nil(t, options)
	assert.ErrorIs(t, err, auth.ErrUserNotFound)
}

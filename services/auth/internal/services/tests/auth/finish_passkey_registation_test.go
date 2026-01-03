package auth_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/gateways"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFinishPasskeyRegistrationSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	fakeUser := helpers.MockUser()
	fakeRawCredential := []byte(gofakeit.Sentence(3))
	fakePasskeySession := helpers.MockPasskeySession()
	fakePasskeyCred := helpers.MockPasskeyCredential()
	serializedPasskeySession, err := json.Marshal(fakePasskeySession)
	require.NoError(t, err)
	authSuite.mockKeyValueStorage.EXPECT().Get(fmt.Sprintf("passkey_session:%d", fakeUser.Id)).
		Return(string(serializedPasskeySession), nil)
	authSuite.mockWebAuthnProvider.EXPECT().VerifyRegistrationOptions(fakeUser.Id, fakeRawCredential, fakePasskeySession).
		Return(fakePasskeyCred, nil)
	authSuite.mockUsersRepo.EXPECT().AddPasskeyCredential(ctx, fakeUser.Id, fakePasskeyCred).Return(nil)

	err = authSuite.service.FinishPasskeyRegistration(ctx, fakeUser.Id, fakeRawCredential)
	assert.NoError(t, err)
}

func TestFinishPasskeyRegistrationInvalidCredential(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	fakeUser := helpers.MockUser()
	fakeRawCredential := []byte(gofakeit.Sentence(3))
	fakePasskeySession := helpers.MockPasskeySession()
	serializedPasskeySession, err := json.Marshal(fakePasskeySession)
	require.NoError(t, err)
	authSuite.mockKeyValueStorage.EXPECT().Get(fmt.Sprintf("passkey_session:%d", fakeUser.Id)).
		Return(string(serializedPasskeySession), nil)
	authSuite.mockWebAuthnProvider.EXPECT().VerifyRegistrationOptions(fakeUser.Id, fakeRawCredential, fakePasskeySession).
		Return(nil, gateways.ErrInvalidCredential)

	err = authSuite.service.FinishPasskeyRegistration(ctx, fakeUser.Id, fakeRawCredential)
	assert.ErrorIs(t, err, auth.ErrInvalidPasskeyCredential)
}

func TestFinishPasskeyRegistrationNotFoundSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	fakeUser := helpers.MockUser()
	fakeRawCredential := []byte(gofakeit.Sentence(3))
	authSuite.mockKeyValueStorage.EXPECT().Get(fmt.Sprintf("passkey_session:%d", fakeUser.Id)).
		Return("", storage.ErrNotFound)

	err := authSuite.service.FinishPasskeyRegistration(ctx, fakeUser.Id, fakeRawCredential)
	assert.ErrorIs(t, err, auth.ErrPasskeyRegistrationNotInProgress)
}

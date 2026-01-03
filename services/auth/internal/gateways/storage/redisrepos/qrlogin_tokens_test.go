package redisrepos_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage/redisrepos"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateQRLoginTokenWithTTL(t *testing.T) {
	testSuite := redisrepos.NewTestSuite(t)
	ctx := context.Background()
	expectedToken := helpers.MockLoginToken()
	expectedTTL := time.Minute

	err := testSuite.QRLoginTokens.CreateWithTTL(ctx, expectedToken, expectedTTL)

	actualTTL, err := testSuite.RedisClient.TTL(ctx, fmt.Sprintf("qrlogin:%s:%s", expectedToken.ClientId, expectedToken.Value)).Result()
	assert.NoError(t, err)
	assert.Equal(t, expectedTTL, actualTTL)
	foundToken, err := testSuite.QRLoginTokens.FindOne(ctx, expectedToken.Value, expectedToken.ClientId)
	require.NoError(t, err)
	assert.Equal(t, expectedToken.Value, foundToken.Value)
	assert.Equal(t, expectedToken.ClientId, foundToken.ClientId)
	assert.Equal(t, expectedToken.IpAddr, foundToken.IpAddr)
	assert.Equal(t, expectedToken.DeviceInfo, foundToken.DeviceInfo)
}

func TestDeleteQRLoginTokensByClient(t *testing.T) {
	testSuite := redisrepos.NewTestSuite(t)
	ctx := context.Background()
	expectedToken := helpers.MockLoginToken()
	expectedTTL := time.Minute
	err := testSuite.QRLoginTokens.CreateWithTTL(ctx, expectedToken, expectedTTL)
	require.NoError(t, err)

	err = testSuite.QRLoginTokens.DeleteAllByClient(ctx, expectedToken.ClientId)
	assert.NoError(t, err)
	foundToken, err := testSuite.QRLoginTokens.FindOne(ctx, expectedToken.Value, expectedToken.ClientId)
	assert.ErrorIs(t, err, storage.ErrNotFound)
	assert.Empty(t, foundToken)
}

package redisrepos_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage/redisrepos"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAuthSession(t *testing.T) {
	testSuite := redisrepos.NewTestSuite(t)
	ctx := context.Background()
	expectedSession := helpers.MockAuthSession()
	expectedTTL := time.Minute

	newSession, err := testSuite.AuthSessions.CreateWithTTL(ctx, expectedSession, expectedTTL)

	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), newSession.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), newSession.LastSeenAt, time.Second)
	assert.Equal(t, expectedSession.Id, newSession.Id)
	actualTTL, err := testSuite.RedisClient.TTL(ctx, fmt.Sprintf("auth-sessions:%d:%s", expectedSession.UserId, expectedSession.Id)).Result()
	require.NoError(t, err)
	assert.Equal(t, expectedTTL, actualTTL)
}

func TestGetAuthSession(t *testing.T) {
	testSuite := redisrepos.NewTestSuite(t)
	ctx := context.Background()
	userId := gofakeit.Number(1, 1000)
	session1 := helpers.MockAuthSession()
	session1.UserId = userId
	session1, err := testSuite.AuthSessions.CreateWithTTL(ctx, session1, time.Minute)
	require.NoError(t, err)
	session2 := helpers.MockAuthSession()
	session2.UserId = userId
	session2, err = testSuite.AuthSessions.CreateWithTTL(ctx, session2, time.Minute)
	require.NoError(t, err)

	t.Run("one by id", func(t *testing.T) {
		foundSession, err := testSuite.AuthSessions.GetById(ctx, session1.UserId, session1.Id)
		assert.NoError(t, err)
		assert.Equal(t, session1, foundSession)
	})

	t.Run("one by login data", func(t *testing.T) {
		foundSession, err := testSuite.AuthSessions.GetByLoginData(ctx, session2.UserId, session2.IpAddr, session2.DeviceInfo)
		assert.NoError(t, err)
		assert.Equal(t, session2, foundSession)
	})

	t.Run("all by user id", func(t *testing.T) {
		foundSessions, err := testSuite.AuthSessions.GetAllByUserId(ctx, session2.UserId)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(foundSessions))
		assert.Contains(t, foundSessions, *session1)
		assert.Contains(t, foundSessions, *session2)
	})
}

func TestUpdateAuthSession(t *testing.T) {
	testSuite := redisrepos.NewTestSuite(t)
	ctx := context.Background()
	expectedSession, err := testSuite.AuthSessions.CreateWithTTL(ctx, helpers.MockAuthSession(), time.Minute)
	require.NoError(t, err)
	expectedLastSeenAt := time.Now().Add(-24 * time.Hour)
	expectedTTL := 5 * time.Minute

	err = testSuite.AuthSessions.UpdateById(ctx, expectedSession.UserId, expectedSession.Id, expectedLastSeenAt, expectedTTL)
	require.NoError(t, err)
	foundSession, err := testSuite.AuthSessions.GetById(ctx, expectedSession.UserId, expectedSession.Id)
	require.NoError(t, err)
	assert.Equal(t, expectedLastSeenAt.Round(time.Second), foundSession.LastSeenAt.Round(time.Second))
	actualTTL, err := testSuite.RedisClient.TTL(ctx, fmt.Sprintf("auth-sessions:%d:%s", expectedSession.UserId, expectedSession.Id)).Result()
	require.NoError(t, err)
	assert.Equal(t, expectedTTL, actualTTL)
}

func TestDeleteAuthSession(t *testing.T) {
	testSuite := redisrepos.NewTestSuite(t)
	ctx := context.Background()

	t.Run("one by session id", func(t *testing.T) {
		session, err := testSuite.AuthSessions.CreateWithTTL(ctx, helpers.MockAuthSession(), time.Minute)
		require.NoError(t, err)

		err = testSuite.AuthSessions.DeleteById(ctx, session.UserId, session.Id)

		require.NoError(t, err)
		_, err = testSuite.AuthSessions.GetById(ctx, session.UserId, session.Id)
		assert.ErrorIs(t, err, storage.ErrNotFound)
	})

	t.Run("all by user id", func(t *testing.T) {
		userId := gofakeit.Number(1, 1000)
		includedSession1 := helpers.MockAuthSession()
		includedSession1.UserId = userId
		includedSession1, err := testSuite.AuthSessions.CreateWithTTL(ctx, includedSession1, time.Minute)
		require.NoError(t, err)
		includedSession2 := helpers.MockAuthSession()
		includedSession2.UserId = userId
		includedSession2, err = testSuite.AuthSessions.CreateWithTTL(ctx, includedSession2, time.Minute)
		require.NoError(t, err)
		excludedSession := helpers.MockAuthSession()
		excludedSession.UserId = userId
		excludedSession, err = testSuite.AuthSessions.CreateWithTTL(ctx, excludedSession, time.Minute)
		require.NoError(t, err)

		err = testSuite.AuthSessions.DeleteAllByUserId(ctx, userId, excludedSession.Id)

		require.NoError(t, err)
		_, err = testSuite.AuthSessions.GetById(ctx, includedSession1.UserId, includedSession1.Id)
		assert.ErrorIs(t, err, storage.ErrNotFound)
		_, err = testSuite.AuthSessions.GetById(ctx, includedSession2.UserId, includedSession2.Id)
		assert.ErrorIs(t, err, storage.ErrNotFound)
		_, err = testSuite.AuthSessions.GetById(ctx, excludedSession.UserId, excludedSession.Id)
		assert.NoError(t, err)
	})
}

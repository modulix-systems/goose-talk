package postgres_repos_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertUser(t *testing.T) {
	repos, ctx := newTestSuite(t)
	user := helpers.MockUser()

	t.Run("success", func(t *testing.T) {
		insertedUser, err := repos.UsersRepo.Insert(ctx, user)
		assert.NoError(t, err)
		assert.NotEmpty(t, insertedUser.ID)
		assert.WithinDuration(t, time.Now(), insertedUser.CreatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), insertedUser.UpdatedAt, time.Second)
		assert.Equal(t, user.Email, insertedUser.Email)
		assert.Equal(t, user.IsActive, insertedUser.IsActive)
	})

	t.Run("already exists", func(t *testing.T) {
		duplicateUser, err := repos.UsersRepo.Insert(ctx, user)
		assert.Nil(t, duplicateUser)
		assert.ErrorIs(t, err, storage.ErrAlreadyExists)
	})
}

func TestCheckExistsWithEmail(t *testing.T) {
	repos, ctx := newTestSuite(t)

	t.Run("false", func(t *testing.T) {
		isExists, err := repos.UsersRepo.CheckExistsWithEmail(ctx, gofakeit.Email())
		assert.NoError(t, err)
		assert.False(t, isExists)
	})
	t.Run("true", func(t *testing.T) {
		user := helpers.MockUser()
		_, err := repos.UsersRepo.Insert(ctx, user)
		require.NoError(t, err)
		isExists, err := repos.UsersRepo.CheckExistsWithEmail(ctx, user.Email)
		assert.NoError(t, err)
		assert.True(t, isExists)
	})
}

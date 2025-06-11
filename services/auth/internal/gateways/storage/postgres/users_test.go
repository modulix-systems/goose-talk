package postgres_repos_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: insert two factor auth entity too if it's supplied along with user
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

func TestGetByLogin(t *testing.T) {
	repos, ctx := newTestSuite(t)
	expectedUser, err := repos.UsersRepo.Insert(ctx, helpers.MockUser())
	require.NoError(t, err)

	assertSuccess := func(user *entity.User, err error) {
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.NotNil(t, user.TwoFactorAuth)
	}

	t.Run("by username", func(t *testing.T) {
		user, err := repos.UsersRepo.GetByLogin(ctx, expectedUser.Username)
		assertSuccess(user, err)
		assert.Equal(t, expectedUser.Username, user.Username)
	})

	t.Run("by email", func(t *testing.T) {
		user, err := repos.UsersRepo.GetByLogin(ctx, expectedUser.Email)
		assertSuccess(user, err)
		assert.Equal(t, expectedUser.Email, user.Email)
	})
	t.Run("not found", func(t *testing.T) {
		user, err := repos.UsersRepo.GetByLogin(ctx, "not found")
		assert.ErrorIs(t, err, storage.ErrNotFound)
		assert.Nil(t, user)
	})
}

func TestGetByID(t *testing.T) {
	repos, ctx := newTestSuite(t)
	expectedUser, err := repos.UsersRepo.Insert(ctx, helpers.MockUser())
	require.NoError(t, err)
	t.Run("success", func(t *testing.T) {
		user, err := repos.UsersRepo.GetByID(ctx, expectedUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.NotNil(t, user.TwoFactorAuth)
	})
	t.Run("not found", func(t *testing.T) {
		user, err := repos.UsersRepo.GetByID(ctx, -1)
		assert.ErrorIs(t, err, storage.ErrNotFound)
		assert.Nil(t, user)
	})
}

func TestUpdateIsActiveById(t *testing.T) {
	repos, ctx := newTestSuite(t)
	expectedUser, err := repos.UsersRepo.Insert(ctx, helpers.MockUser())
	require.NoError(t, err)
	expectedIsActive := gofakeit.Bool()
	t.Run("success", func(t *testing.T) {
		user, err := repos.UsersRepo.UpdateIsActiveById(ctx, expectedUser.ID, expectedIsActive)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedIsActive, user.IsActive)
	})
	t.Run("not found", func(t *testing.T) {
		user, err := repos.UsersRepo.UpdateIsActiveById(ctx, -1, expectedIsActive)
		assert.ErrorIs(t, err, storage.ErrNotFound)
		assert.Nil(t, user)
	})
}

// func TestAddPasskeyCredential(t *testing.T) {
// 	repos, ctx := newTestSuite(t)
// 	expectedUser, err := repos.UsersRepo.Insert(ctx, helpers.MockUser())
// 	require.NoError(t, err)
// 	expectedCredential := helpers.MockPasskeyCredential()
// 	t.Run("success", func(t *testing.T) {
// 		user, err := repos.UsersRepo.UpdateIsActiveById(ctx, expectedUser.ID, expectedIsActive)
// 		assert.NoError(t, err)
// 		assert.Equal(t, expectedUser.ID, user.ID)
// 		assert.Equal(t, expectedIsActive, user.IsActive)
// 	})
// 	t.Run("user not found", func(t *testing.T) {
// 		user, err := repos.UsersRepo.AddPasskeyCredential(ctx, -1, expectedIsActive)
// 		assert.ErrorIs(t, err, storage.ErrNotFound)
// 		assert.Nil(t, user)
// 	})
// }

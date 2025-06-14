package postgres_repos_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	postgres_repos "github.com/modulix-systems/goose-talk/internal/gateways/storage/postgres"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertUser(t *testing.T) {
	repos, ctx := newTestSuite(t)

	checkUserValid := func(expectedUser *entity.User, insertedUser *entity.User, err error) {
		assert.NoError(t, err)
		assert.NotEmpty(t, insertedUser.ID)
		assert.WithinDuration(t, time.Now(), insertedUser.CreatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), insertedUser.UpdatedAt, time.Second)
		assert.Equal(t, expectedUser.Email, insertedUser.Email)
		assert.Equal(t, expectedUser.IsActive, insertedUser.IsActive)
	}

	checkUserInDB := func(expectedUser *entity.User, insertedID int) *entity.User {
		userFromDB, err := repos.UsersRepo.GetByID(ctx, insertedID)
		require.NoError(t, err)
		checkUserValid(expectedUser, userFromDB, err)
		return userFromDB
	}

	t.Run("success", func(t *testing.T) {
		user := helpers.MockUser()
		user.TwoFactorAuth = nil
		insertedUser, err := repos.UsersRepo.Insert(ctx, user)
		checkUserValid(user, insertedUser, err)
		checkUserInDB(user, insertedUser.ID)
	})

	t.Run("insert along with 2 fa related entity", func(t *testing.T) {
		user := helpers.MockUser()
		insertedUser, err := repos.UsersRepo.Insert(ctx, user)
		checkUserValid(user, insertedUser, err)
		require.NotNil(t, insertedUser.TwoFactorAuth)
		assert.Equal(t, insertedUser.ID, insertedUser.TwoFactorAuth.UserId)
		assert.Equal(t, user.TwoFactorAuth.Transport, insertedUser.TwoFactorAuth.Transport)
	})

	t.Run("already exists", func(t *testing.T) {
		user := helpers.MockUser()
		user.TwoFactorAuth = nil
		_, err := repos.UsersRepo.Insert(ctx, user)
		require.NoError(t, err)
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
		require.NotNil(t, user.TwoFactorAuth)
		assert.Equal(t, user.ID, user.TwoFactorAuth.UserId)
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

func assertUserHasCred(t *testing.T, credsList []entity.PasskeyCredential, expectedCredential *entity.PasskeyCredential) {
	t.Helper()
	var userHasCred bool
	for _, cred := range credsList {
		if cred.ID == expectedCredential.ID {
			userHasCred = true
			assert.Equal(t, cred.UserId, expectedCredential.UserId)
		}
	}
	assert.True(t, userHasCred)
}

func TestAddPasskeyCredential(t *testing.T) {
	repos, ctx := newTestSuite(t)
	expectedUser, err := repos.UsersRepo.Insert(ctx, helpers.MockUser())
	require.NoError(t, err)
	expectedCredential := helpers.MockPasskeyCredential()
	expectedCredential.UserId = expectedUser.ID
	t.Run("success", func(t *testing.T) {
		err := repos.UsersRepo.AddPasskeyCredential(ctx, expectedUser.ID, expectedCredential)
		assert.NoError(t, err)
		user, err := repos.UsersRepo.GetByIDWithPasskeyCredentials(ctx, expectedUser.ID)
		require.NoError(t, err)
		require.NotNil(t, user)
		assertUserHasCred(t, user.PasskeyCredentials, expectedCredential)
	})
	t.Run("user not found", func(t *testing.T) {
		// change id to avoid unique violation
		expectedCredential.ID = gofakeit.UUID()
		err := repos.UsersRepo.AddPasskeyCredential(ctx, -1, expectedCredential)
		assert.ErrorIs(t, err, storage.ErrNotFound)
	})
}

func TestGetByIDWithPasskeyCredentials(t *testing.T) {
	repos, ctx := newTestSuite(t)
	expectedUser, err := repos.UsersRepo.Insert(ctx, helpers.MockUser())
	require.NoError(t, err)
	expectedCredential := helpers.MockPasskeyCredential()
	expectedCredential.UserId = expectedUser.ID
	qb := repos.UsersRepo.Builder.Insert(`"passkey_credential"(id, public_key, user_id, transports)`).
		Values(expectedCredential.ID, expectedCredential.PublicKey, expectedCredential.UserId, expectedCredential.Transports)
	queryable, err := postgres_repos.GetQueryable(ctx, postgres_repos.PgxPoolAdapter{repos.UsersRepo.Pool})
	require.NoError(t, err)
	query, args := qb.MustSql()
	_, err = queryable.Exec(ctx, query, args...)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		user, err := repos.UsersRepo.GetByIDWithPasskeyCredentials(ctx, expectedUser.ID)
		assert.NoError(t, err)
		require.NotNil(t, user)
		assertUserHasCred(t, user.PasskeyCredentials, expectedCredential)
	})

	t.Run("success no credentials", func(t *testing.T) {
		expectedUser, err := repos.UsersRepo.Insert(ctx, helpers.MockUser())
		require.NoError(t, err)
		user, err := repos.UsersRepo.GetByIDWithPasskeyCredentials(ctx, expectedUser.ID)
		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Nil(t, user.PasskeyCredentials)
	})

	t.Run("not found user", func(t *testing.T) {
		user, err := repos.UsersRepo.GetByIDWithPasskeyCredentials(ctx, -1)
		assert.ErrorIs(t, err, storage.ErrNotFound)
		assert.Nil(t, user)
	})
}

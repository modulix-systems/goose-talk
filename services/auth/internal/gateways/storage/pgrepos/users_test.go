package pgrepos_test

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage/pgrepos"
	"github.com/modulix-systems/goose-talk/pkg/postgres"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertUser(t *testing.T) {
	testSuite := pgrepos.NewTestSuite(t)

	checkUserValid := func(expectedUser *entity.User, insertedUser *entity.User) {
		assert.NotEmpty(t, insertedUser.ID)
		assert.WithinDuration(t, time.Now(), insertedUser.CreatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), insertedUser.UpdatedAt, time.Second)
		assert.Equal(t, expectedUser.Email, insertedUser.Email)
		assert.Equal(t, expectedUser.IsActive, insertedUser.IsActive)
	}

	checkUserInDB := func(expectedUser *entity.User, insertedID int) *entity.User {
		userFromDB, err := testSuite.Users.GetByID(testSuite.TxCtx, insertedID)
		require.NoError(t, err)
		checkUserValid(expectedUser, userFromDB)
		return userFromDB
	}

	t.Run("success", func(t *testing.T) {
		user := helpers.MockUser()
		user.TwoFactorAuth = nil
		insertedUser, err := testSuite.Users.Insert(testSuite.TxCtx, user)
		require.NoError(t, err)
		checkUserValid(user, insertedUser)
		checkUserInDB(user, insertedUser.ID)
	})

	t.Run("insert along with 2 fa related entity", func(t *testing.T) {
		user := helpers.MockUser()
		insertedUser, err := testSuite.Users.Insert(testSuite.TxCtx, user)
		require.NoError(t, err)
		checkUserValid(user, insertedUser)
		require.NotNil(t, insertedUser.TwoFactorAuth)
		assert.Equal(t, insertedUser.ID, insertedUser.TwoFactorAuth.UserId)
		assert.Equal(t, user.TwoFactorAuth.Transport, insertedUser.TwoFactorAuth.Transport)
	})

	t.Run("already exists", func(t *testing.T) {
		user := helpers.MockUser()
		user.TwoFactorAuth = nil
		_, err := testSuite.Users.Insert(testSuite.TxCtx, user)
		require.NoError(t, err)
		duplicateUser, err := testSuite.Users.Insert(testSuite.TxCtx, user)
		assert.Nil(t, duplicateUser)
		assert.ErrorIs(t, err, storage.ErrAlreadyExists)
	})
}

func TestCheckExistsWithEmail(t *testing.T) {
	testSuite := pgrepos.NewTestSuite(t)

	t.Run("false", func(t *testing.T) {
		isExists, err := testSuite.Users.CheckExistsWithEmail(testSuite.TxCtx, gofakeit.Email())
		assert.NoError(t, err)
		assert.False(t, isExists)
	})
	t.Run("true", func(t *testing.T) {
		user := helpers.MockUser()
		_, err := testSuite.Users.Insert(testSuite.TxCtx, user)
		require.NoError(t, err)
		isExists, err := testSuite.Users.CheckExistsWithEmail(testSuite.TxCtx, user.Email)
		assert.NoError(t, err)
		assert.True(t, isExists)
	})
}

func TestGetByLogin(t *testing.T) {
	testSuite := pgrepos.NewTestSuite(t)
	expectedUser, err := testSuite.Users.Insert(testSuite.TxCtx, helpers.MockUser())
	require.NoError(t, err)

	assertSuccess := func(user *entity.User, err error) {
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		require.NotNil(t, user.TwoFactorAuth)
		assert.Equal(t, user.ID, user.TwoFactorAuth.UserId)
	}

	t.Run("by username", func(t *testing.T) {
		user, err := testSuite.Users.GetByLogin(testSuite.TxCtx, expectedUser.Username)
		assertSuccess(user, err)
		assert.Equal(t, expectedUser.Username, user.Username)
	})

	t.Run("by email", func(t *testing.T) {
		user, err := testSuite.Users.GetByLogin(testSuite.TxCtx, expectedUser.Email)
		assertSuccess(user, err)
		assert.Equal(t, expectedUser.Email, user.Email)
	})
	t.Run("not found", func(t *testing.T) {
		user, err := testSuite.Users.GetByLogin(testSuite.TxCtx, "not found")
		assert.ErrorIs(t, err, storage.ErrNotFound)
		assert.Nil(t, user)
	})
}

func TestGetByID(t *testing.T) {
	testSuite := pgrepos.NewTestSuite(t)
	expectedUser, err := testSuite.Users.Insert(testSuite.TxCtx, helpers.MockUser())
	require.NoError(t, err)
	t.Run("success", func(t *testing.T) {
		user, err := testSuite.Users.GetByID(testSuite.TxCtx, expectedUser.ID)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.NotNil(t, user.TwoFactorAuth)
	})
	t.Run("not found", func(t *testing.T) {
		user, err := testSuite.Users.GetByID(testSuite.TxCtx, -1)
		assert.ErrorIs(t, err, storage.ErrNotFound)
		assert.Nil(t, user)
	})
}

func TestUpdateIsActiveById(t *testing.T) {
	testSuite := pgrepos.NewTestSuite(t)
	expectedUser, err := testSuite.Users.Insert(testSuite.TxCtx, helpers.MockUser())
	require.NoError(t, err)
	expectedIsActive := gofakeit.Bool()
	t.Run("success", func(t *testing.T) {
		user, err := testSuite.Users.UpdateIsActiveById(testSuite.TxCtx, expectedUser.ID, expectedIsActive)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedIsActive, user.IsActive)
	})
	t.Run("not found", func(t *testing.T) {
		user, err := testSuite.Users.UpdateIsActiveById(testSuite.TxCtx, -1, expectedIsActive)
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

func TestCreatePasskeyCredential(t *testing.T) {
	testSuite := pgrepos.NewTestSuite(t)
	expectedUser, err := testSuite.Users.Insert(testSuite.TxCtx, helpers.MockUser())
	require.NoError(t, err)
	expectedCredential := helpers.MockPasskeyCredential()
	expectedCredential.UserId = expectedUser.ID
	t.Run("success", func(t *testing.T) {
		err := testSuite.Users.CreatePasskeyCredential(testSuite.TxCtx, expectedUser.ID, expectedCredential)
		assert.NoError(t, err)
		user, err := testSuite.Users.GetByIDWithPasskeyCredentials(testSuite.TxCtx, expectedUser.ID)
		require.NoError(t, err)
		require.NotNil(t, user)
		assertUserHasCred(t, user.PasskeyCredentials, expectedCredential)
	})
	t.Run("user not found", func(t *testing.T) {
		// change id to avoid unique violation
		expectedCredential.ID = gofakeit.UUID()
		err := testSuite.Users.CreatePasskeyCredential(testSuite.TxCtx, -1, expectedCredential)
		assert.ErrorIs(t, err, storage.ErrNotFound)
	})
}

func TestGetByIDWithPasskeyCredentials(t *testing.T) {
	testSuite := pgrepos.NewTestSuite(t)
	expectedUser, err := testSuite.Users.Insert(testSuite.TxCtx, helpers.MockUser())
	require.NoError(t, err)
	expectedCredential := helpers.MockPasskeyCredential()
	expectedCredential.UserId = expectedUser.ID
	qb := testSuite.Users.Builder.Insert(`"passkey_credential"(id, public_key, user_id, transports)`).
		Values(expectedCredential.ID, expectedCredential.PublicKey, expectedCredential.UserId, expectedCredential.Transports)
	query, args := qb.MustSql()
	_, err = testSuite.Tx.Exec(testSuite.TxCtx, query, args...)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		user, err := testSuite.Users.GetByIDWithPasskeyCredentials(testSuite.TxCtx, expectedUser.ID)
		assert.NoError(t, err)
		require.NotNil(t, user)
		assertUserHasCred(t, user.PasskeyCredentials, expectedCredential)
	})

	t.Run("success no credentials", func(t *testing.T) {
		expectedUser, err := testSuite.Users.Insert(testSuite.TxCtx, helpers.MockUser())
		require.NoError(t, err)
		user, err := testSuite.Users.GetByIDWithPasskeyCredentials(testSuite.TxCtx, expectedUser.ID)
		assert.NoError(t, err)
		require.NotNil(t, user)
		assert.Nil(t, user.PasskeyCredentials)
	})

	t.Run("not found user", func(t *testing.T) {
		user, err := testSuite.Users.GetByIDWithPasskeyCredentials(testSuite.TxCtx, -1)
		assert.ErrorIs(t, err, storage.ErrNotFound)
		assert.Nil(t, user)
	})
}

func TestCreateTwoFa(t *testing.T) {
	testSuite := pgrepos.NewTestSuite(t)

	mockUser := helpers.MockUser()
	mockTwoFa := *mockUser.TwoFactorAuth
	mockUser.TwoFactorAuth = nil
	mockUser, err := testSuite.Users.Insert(testSuite.TxCtx, mockUser)
	require.NoError(t, err)
	mockTwoFa.UserId = mockUser.ID
	expectedTwoFa, err := testSuite.Users.CreateTwoFa(testSuite.TxCtx, &mockTwoFa)
	require.NoError(t, err)
	require.NotNil(t, expectedTwoFa.UserId)

	actualTwoFa, err := postgres.ExecAndGetOne[entity.TwoFactorAuth](
		testSuite.TxCtx,
		testSuite.Users.Builder.Select("user_id, transport").From("two_factor_auth"),
		testSuite.Pg.Pool,
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, expectedTwoFa.UserId, actualTwoFa.UserId)
	assert.Equal(t, expectedTwoFa.Transport, actualTwoFa.Transport)
}

func TestUpdateTwoFaContact(t *testing.T) {
	testSuite := pgrepos.NewTestSuite(t)

	expectedContact := gofakeit.Username()
	mockUser := helpers.MockUser()
	mockUser, err := testSuite.Users.Insert(testSuite.TxCtx, mockUser)
	require.NoError(t, err)
	require.NotEmpty(t, mockUser.TwoFactorAuth)

	err = testSuite.Users.UpdateTwoFaContact(testSuite.TxCtx, mockUser.TwoFactorAuth.UserId, expectedContact)
	require.NoError(t, err)

	actualTwoFa, err := postgres.ExecAndGetOne[entity.TwoFactorAuth](
		testSuite.TxCtx,
		testSuite.Users.Builder.Select("contact").From("two_factor_auth"),
		testSuite.Pg.Pool,
		nil,
	)
	require.NoError(t, err)
	assert.Equal(t, expectedContact, actualTwoFa.Contact)
}

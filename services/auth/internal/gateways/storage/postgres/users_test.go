package postgres_repos_test

import (
	"context"
	"testing"
	"time"

	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	postgres_repos "github.com/modulix-systems/goose-talk/internal/gateways/storage/postgres"
	"github.com/modulix-systems/goose-talk/internal/services"
	"github.com/modulix-systems/goose-talk/pkg/postgres"
	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
	"github.com/stretchr/testify/assert"
)

func TestInsertUser(t *testing.T) {
	ctx := context.Background()
	pg, tx := postgres.NewTestSuite(t, ctx)
	repos := postgres_repos.NewRepositorories(pg)
	ctx = services.SetTransaction(ctx, tx)
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

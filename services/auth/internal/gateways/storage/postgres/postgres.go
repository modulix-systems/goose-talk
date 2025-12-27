package postgres_repos

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/modulix-systems/goose-talk/internal/config"
	users_repo "github.com/modulix-systems/goose-talk/internal/gateways/storage/postgres/users"
	"github.com/modulix-systems/goose-talk/pkg/postgres"
)

type Repositories struct {
	Users *users_repo.Repository
}

func New(pg *postgres.Postgres) *Repositories {
	return &Repositories{Users: &users_repo.Repository{pg}}
}

type TestSuite struct {
	Repositories
	TxCtx context.Context
	Tx    pgx.Tx
}

func NewTestSuite(t *testing.T) *TestSuite {
	ctx := context.Background()
	pg, tx := postgres.NewTestSuite(t, ctx)
	repos := New(pg)

	return &TestSuite{
		TxCtx:        context.WithValue(ctx, config.TransactionCtxKey, tx),
		Repositories: *repos,
		Tx:           tx,
	}
}

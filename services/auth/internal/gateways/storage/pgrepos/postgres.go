package pgrepos

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/modulix-systems/goose-talk/internal/config"
	"github.com/modulix-systems/goose-talk/pkg/postgres"
)

type Repositories struct {
	Users *UsersRepo
}

func New(pg *postgres.Postgres) *Repositories {
	return &Repositories{Users: &UsersRepo{pg}}
}

type TestSuite struct {
	Repositories
	TxCtx context.Context
	Tx    pgx.Tx
	Pg    *postgres.Postgres
}

func NewTestSuite(t *testing.T) *TestSuite {
	ctx := context.Background()
	pg, tx := postgres.NewTestSuite(t, ctx)
	repos := New(pg)

	return &TestSuite{
		TxCtx:        context.WithValue(ctx, config.TRANSACTION_CTX_KEY, tx),
		Repositories: *repos,
		Tx:           tx,
		Pg:           pg,
	}
}

package postgres

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/modulix-systems/goose-talk/internal/config"
	"github.com/modulix-systems/goose-talk/internal/utils"
)

func NewTestSuite(t *testing.T, ctx context.Context) (*Postgres, pgx.Tx) {
	rootPath := utils.FindRootPath()
	if rootPath == "" {
		t.Fatal("Unable to find root path")
	}
	cfg := config.MustLoad(filepath.Join(rootPath, "configs", "tests.yaml"))
	pg, err := New(cfg.PG.Dsn)
	if err != nil {
		t.Fatal(err)
	}
	txManager := NewTransactionManager(pg.Pool)
	tx, err := txManager.StartTransaction(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := tx.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
	})
	return pg, tx.(pgx.Tx)
}

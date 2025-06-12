//go:generate mockgen -source=postgres.go -destination=../../../../tests/mocks/mocks_postgres.go -package=mocks

package postgres_repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modulix-systems/goose-talk/internal/services"
	"github.com/modulix-systems/goose-talk/pkg/postgres"
)

type pgErrCode string

const (
	UniqueViolationErrCode     pgErrCode = "23505"
	ForeignKeyViolationErrCode pgErrCode = "23503"
)

type Repositories struct {
	UsersRepo *UsersRepo
}

func NewRepositorories(pg *postgres.Postgres) *Repositories {
	return &Repositories{UsersRepo: &UsersRepo{pg}}
}

type Queryable interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type SupportsAcquire interface {
	Acquire(ctx context.Context) (Queryable, error)
}

func GetQueryable(ctx context.Context, connProvider SupportsAcquire) (Queryable, error) {
	tx := ctx.Value(services.TransactionCtxKey)
	if tx != nil {
		return tx.(Queryable), nil
	}
	conn, err := connProvider.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn from pool: %w", err)
	}
	return conn, nil
}

func getPgErrCode(err error) pgErrCode {
	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) {
		return pgErrCode(pgxErr.Code)
	}
	return ""
}

type PgxPoolAdapter struct {
	Pool *pgxpool.Pool
}

func (a PgxPoolAdapter) Acquire(ctx context.Context) (Queryable, error) {
	return a.Pool.Acquire(ctx)
}

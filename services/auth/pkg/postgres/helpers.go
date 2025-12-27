//go:generate mockgen -source=helpers.go -destination=../../tests/mocks/mocks_postgres.go -package=mocks

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modulix-systems/goose-talk/internal/config"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
)

type queryBuilder interface {
	ToSql() (string, []any, error)
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
	tx := ctx.Value(config.TransactionCtxKey)
	if tx != nil {
		return tx.(Queryable), nil
	}
	conn, err := connProvider.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn from pool: %w", err)
	}
	return conn, nil
}

type PgxPoolAdapter struct {
	Pool *pgxpool.Pool
}

func (a PgxPoolAdapter) Acquire(ctx context.Context) (Queryable, error) {
	return a.Pool.Acquire(ctx)
}


func ExecAndGetMany[T any](ctx context.Context, qb queryBuilder, pool *pgxpool.Pool, collectFunc pgx.RowToFunc[T]) ([]T, error) {
	if collectFunc == nil {
		collectFunc = pgx.RowToStructByNameLax[T]
	}
	queryable, err := GetQueryable(ctx, PgxPoolAdapter{pool})
	if err != nil {
		return nil, err
	}
	query, args, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %w", err)
	}
	rows, err := queryable.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}
	res, err := pgx.CollectRows(rows, collectFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to collect row into user struct: %w", err)
	}
	if len(res) == 0 {
		return nil, storage.ErrNotFound
	}
	return res, nil
}

func ExecAndGetOne[T any](ctx context.Context, qb queryBuilder, pool *pgxpool.Pool, collectFunc pgx.RowToFunc[T]) (*T, error) {
	res, err := ExecAndGetMany(ctx, qb, pool, collectFunc)
	if err != nil {
		return nil, err
	}
	return &res[0], nil
}

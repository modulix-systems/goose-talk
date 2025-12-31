//go:generate mockgen -source=helpers.go -destination=../../tests/mocks/mocks_postgres.go -package=mocks

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

func GetQueryable(ctx context.Context, acquirable SupportsAcquire) (Queryable, error) {
	tx := ctx.Value(config.TRANSACTION_CTX_KEY)
	if tx != nil {
		return tx.(Queryable), nil
	}
	conn, err := acquirable.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn from pool: %w", err)
	}
	return conn, nil
}

func ExecAndGetMany[T any](ctx context.Context, qb queryBuilder, pool *PGPool, mapper pgx.RowToFunc[T]) ([]T, error) {
	if mapper == nil {
		mapper = pgx.RowToStructByNameLax[T]
	}
	queryable, err := GetQueryable(ctx, pool)
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
	res, err := pgx.CollectRows(rows, mapper)
	if err != nil {
		return nil, fmt.Errorf("failed to collect row into user struct: %w", err)
	}
	if len(res) == 0 {
		return nil, storage.ErrNotFound
	}
	return res, nil
}

func ExecAndGetOne[T any](ctx context.Context, qb queryBuilder, pool *PGPool, mapper pgx.RowToFunc[T]) (*T, error) {
	res, err := ExecAndGetMany(ctx, qb, pool, mapper)
	if err != nil {
		return nil, err
	}
	return &res[0], nil
}

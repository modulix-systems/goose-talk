//go:generate mockgen -source=helpers.go -destination=./postgres_mocks_test.go -package=postgres_test

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type queryBuilder interface {
	ToSql() (string, []any, error)
}

type Queryable interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Releaseable interface {
	Release()
}

type Acquirable interface {
	Acquire(ctx context.Context) (Queryable, error)
}

func GetQueryable(ctx context.Context, acquirable Acquirable, transactionCtxKey string) (Queryable, error) {
	tx := ctx.Value(transactionCtxKey)
	if tx != nil {
		return tx.(Queryable), nil
	}
	fmt.Println("Acquire new connection")
	conn, err := acquirable.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire conn from pool: %w", err)
	}
	return conn, nil
}

func ExecAndGetMany[T any](ctx context.Context, qb queryBuilder, pool *PGPool, mapper pgx.RowToFunc[T], transactionCtxKey string) ([]T, error) {
	if mapper == nil {
		mapper = pgx.RowToStructByNameLax[T]
	}
	queryable, err := GetQueryable(ctx, pool, transactionCtxKey)
	if err != nil {
		return nil, err
	}
	if r, ok := queryable.(Releaseable); ok {
		defer r.Release()
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
	return res, nil
}

func ExecAndGetOne[T any](ctx context.Context, qb queryBuilder, pool *PGPool, mapper pgx.RowToFunc[T], transactionCtxKey string) (*T, error) {
	res, err := ExecAndGetMany(ctx, qb, pool, mapper, transactionCtxKey)
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, ErrNoRows
	}
	return &res[0], nil
}

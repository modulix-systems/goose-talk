package postgres_repos

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
)

type queryBuilder interface {
	ToSql() (string, []any, error)
}

func execAndCollectRows[T any](ctx context.Context, qb queryBuilder, pool *pgxpool.Pool) ([]T, error) {
	queryable, err := GetQueryable(ctx, pgxPoolAdapter{pool})
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
	res, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[T])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row into user struct: %w", err)
	}
	if len(res) == 0 {
		return nil, storage.ErrNotFound
	}
	return res, nil
}

func execAndGetOne[T any](ctx context.Context, qb queryBuilder, pool *pgxpool.Pool) (*T, error) {
	res, err := execAndCollectRows[T](ctx, qb, pool)
	if err != nil {
		return nil, err
	}
	return &res[0], nil
}

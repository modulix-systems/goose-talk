package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type pgConn struct {
	Queryable
}

func (c *pgConn) Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error) {
	commandTag, err = c.Queryable.Exec(ctx, sql, arguments...)
	if err != nil {
		return pgconn.CommandTag{}, mapPgxError(err)
	}
	return
}

func (c *pgConn) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	rows, err := c.Queryable.Query(ctx, sql, args...)
	if err != nil {
		return nil, mapPgxError(err)
	}
	return rows, nil
}

func (c *pgConn) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return c.Queryable.QueryRow(ctx, sql, args...)
}

func newPgQueryable(base Queryable) Queryable {
	return &pgConn{base}
}
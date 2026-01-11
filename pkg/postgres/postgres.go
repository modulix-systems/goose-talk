// Package postgres implements postgres connection.
package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_defaultMaxPoolSize       = 1
	_defaultConnAttempts      = 10
	_defaultConnTimeout       = time.Second
	_defaultTransactionCtxKey = "pg-transaction"
)

// Postgres -.
type Postgres struct {
	maxPoolSize       int
	connAttempts      int
	connTimeout       time.Duration
	TransactionCtxKey string

	Builder squirrel.StatementBuilderType
	Pool    *PGPool
}

// New -.
func New(dsn string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:       _defaultMaxPoolSize,
		connAttempts:      _defaultConnAttempts,
		connTimeout:       _defaultConnTimeout,
		TransactionCtxKey: _defaultTransactionCtxKey,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres - New - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)

	for pg.connAttempts > 0 {
		var pool *pgxpool.Pool
		pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err == nil {
			pg.Pool = &PGPool{pool}
			break
		}

		log.Printf("Postgres is trying to connect, attempts left: %d", pg.connAttempts)

		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - connAttempts == 0: %w", err)
	}

	return pg, nil
}

// Close -.
func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

type pgTransactionManager struct {
	pool *PGPool
}

func NewTransactionManager(pool *PGPool) *pgTransactionManager {
	return &pgTransactionManager{
		pool: pool,
	}
}

type QueryableTransaction interface {
	Queryable
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

func (m *pgTransactionManager) StartTransaction(ctx context.Context) (QueryableTransaction, error) {
	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	return newPgQueryable(tx).(QueryableTransaction), nil
}

type PGPool struct {
	*pgxpool.Pool
}

func (p *PGPool) Acquire(ctx context.Context) (Queryable, error) {
	conn, err := p.Pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	return &pgConn{conn}, nil
}

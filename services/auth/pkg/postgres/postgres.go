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
	"github.com/modulix-systems/goose-talk/internal/gateways"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

// Postgres -.
type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	Builder squirrel.StatementBuilderType
	Pool    *PGPool
}

// New -.
func New(dsn string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
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

func (m *pgTransactionManager) StartTransaction(ctx context.Context) (gateways.Transaction, error) {
	return m.pool.BeginTx(ctx, pgx.TxOptions{})
}

type PGPool struct {
	*pgxpool.Pool
}

func (p *PGPool) Acquire(ctx context.Context) (Queryable, error) {
	return p.Pool.Acquire(ctx)
}

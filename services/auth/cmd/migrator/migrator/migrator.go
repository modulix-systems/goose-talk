package migrator

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/lib/pq"
)

type Migrator struct {
	sourceDrv   source.Driver
	databaseDrv database.Driver
	db          *sql.DB
	schemaName  string
	migrationsTableName string
	*migrate.Migrate
}

func New(sourceURL, databaseURL string) (*Migrator, error) {
	sourceDrv, err := source.Open(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open source, %q: %w", sourceURL, err)
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	cfg := &postgres.Config{}
	databaseDrv, err := postgres.WithInstance(db, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	migrator, err := migrate.NewWithInstance("", sourceDrv, "", databaseDrv)
	if err != nil {
		return nil, err
	}

	return &Migrator{sourceDrv, databaseDrv, db, cfg.SchemaName, cfg.MigrationsTable, migrator}, nil
}

func (m *Migrator) Drop() error {
	if _, err := m.db.Exec(fmt.Sprintf("DROP SCHEMA %s CASCADE", m.schemaName)); err != nil {
		return err
	}
	if _, err := m.db.Exec(fmt.Sprintf("CREATE SCHEMA %s", m.schemaName)); err != nil {
		return err
	}

	query := `CREATE TABLE IF NOT EXISTS ` + pq.QuoteIdentifier(m.schemaName) + `.` + pq.QuoteIdentifier(m.migrationsTableName) + ` (version bigint not null primary key, dirty boolean not null)`
	if _, err := m.db.Exec(query); err != nil {
		return &database.Error{OrigErr: err, Query: []byte(query)}
	}

	return nil
}

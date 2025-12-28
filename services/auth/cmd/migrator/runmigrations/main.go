package runmigrations

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/modulix-systems/goose-talk/internal/config"
)

func Exec(migrationsPath string) {
	var resetDB bool
	flag.BoolVar(&resetDB, "reset", false, "specify if you want to reset the database before running migration")
	os.Args = os.Args[2:]
	flag.Parse()

	cfgPath := config.ResolveConfigPath()
	cfg := config.MustLoad(cfgPath)

	m, err := migrate.New("file://"+migrationsPath, cfg.Postgres.Url)
	if err != nil {
		panic(err)
	}

	if resetDB {
		fmt.Println("Resetting database...")
		if err := m.Drop(); err != nil {
			panic(err)
		}
		fmt.Println("Database reset")
	}

	var direction string
	if len(os.Args) > 0 {
		direction = os.Args[0]
	} else {
		direction = "up"
	}
	switch strings.ToLower(direction) {
	case "up":
		if migrated := migrateUp(m); !migrated {
			return
		}
	case "down":
		var steps string
		if len(os.Args) > 0 {
			steps = os.Args[0]
		}
		if migrated := migrateDown(steps, m); !migrated {
			return
		}
	default:
		panic(fmt.Sprintf("Unknown direction: %s", direction))
	}
	fmt.Println("Migrations applied successfully")
}

func migrateUp(m *migrate.Migrate) (migrated bool) {
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("No migrations to apply")
			return
		}

		var dirtyErr migrate.ErrDirty
		if errors.As(err, &dirtyErr) {
			tryRecoverDirtyMigration(dirtyErr, m)
		}

		panic(err)
	}
	migrated = true
	return
}

func migrateDown(steps string, m *migrate.Migrate) (migrated bool) {
	if steps != "" {
		stepsInt, err := strconv.Atoi(steps)
		if err != nil {
			panic(err)
		}
		if err := m.Steps(-stepsInt); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("No migrations to apply")
				return
			}

			var dirtyErr migrate.ErrDirty
			if errors.As(err, &dirtyErr) {
				tryRecoverDirtyMigration(dirtyErr, m)
			}

			panic(err)
		}
		migrated = true
		return
	}
	// Ask confirmation to downgrade all migrations if steps not provided
	fmt.Print("Are you sure that you want to downgrade all migrations? (yes/no) [no] ")
	confirmed, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		panic(err)
	}
	switch strings.ToLower(confirmed) {
	case "\n", "no":
		fmt.Println("Aborting...")
	case "yes":
		if err = m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("No migrations to apply")
				return
			}

			var dirtyErr migrate.ErrDirty
			if errors.As(err, &dirtyErr) {
				tryRecoverDirtyMigration(dirtyErr, m)
			}

			return
		}
		migrated = true
	default:
		panic(fmt.Sprintf("Invalid value: %s. Choices are: %s", confirmed, "yes/no"))
	}
	return
}

func tryRecoverDirtyMigration(dirtyErr migrate.ErrDirty, m *migrate.Migrate) {
	if err := m.Force(dirtyErr.Version - 1); err != nil {
		panic(fmt.Errorf("Error during recovering dirty error: %w", err))
	}
}

package main

import (
	"flag"
	"os"
	"path"
	"strings"

	"github.com/modulix-systems/goose-talk/migrator/internal/cli/commands"
)

func main() {
	var migrationsPath, databaseUrl string
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&databaseUrl, "db-url", "", "database URI")
	if migrationsPath == "" {
		currDir, err := os.Getwd()
		if err != nil {
			panic("Unable to resolve migrations location")
		}
		migrationsPath = path.Join(currDir, "migrations")
		if _, err := os.Stat(migrationsPath); err != nil {
			panic(err)
		}
	}
	if databaseUrl == "" {
		databaseUrl = os.Getenv("DB_URL")
	}

	if len(os.Args) < 2 {
		panic("Invalid command. Options are: migrate, make-migrations")
	}

	subcommand := strings.ToLower(os.Args[1])
	os.Args = os.Args[2:]
	switch subcommand {
	case "migrate":
		if databaseUrl == "" {
			panic("Can't run migrations, database url is not specified!")
		}
		commands.RunMigrations(migrationsPath, databaseUrl)
	case "make-migrations":
		commands.MakeMigrations(migrationsPath)
	}
}

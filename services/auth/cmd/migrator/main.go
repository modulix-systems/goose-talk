package main

import (
	"flag"
	"os"
	"path"
	"strings"

	"github.com/modulix-systems/goose-talk/cmd/migrator/makemigrations"
	"github.com/modulix-systems/goose-talk/cmd/migrator/runmigrations"
)

func main() {
	var migrationsPath string
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
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

	if len(os.Args) < 2 {
		panic("Invalid command. Options are: migrate, make-migrations")
	}

	subcommand := strings.ToLower(os.Args[1])
	if subcommand == "migrate" {
		runmigrations.Exec(migrationsPath)
		return
	}
	if subcommand == "make-migrations" {
		makemigrations.Exec(migrationsPath)
		return
	}
}

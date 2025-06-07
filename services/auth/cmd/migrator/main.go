package main

import (
	"flag"
	"os"
	"path"

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
	switch os.Args[1] {
	case "migrate":
		runmigrations.Exec(migrationsPath)
	case "make-migrations":
		makemigrations.Exec(migrationsPath)
	default:
		panic("Invalid command. Options are: migrate, make-migrations")
	}
}

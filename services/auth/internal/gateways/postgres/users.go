package postgres_repos

import "github.com/modulix-systems/goose-talk/pkg/postgres"

type UsersRepo struct {
	*postgres.Postgres
}

func (repo *UsersRepo) Insert() {}

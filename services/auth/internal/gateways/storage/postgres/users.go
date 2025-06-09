package postgres_repos

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/pkg/postgres"
)

type UsersRepo struct {
	*postgres.Postgres
}

func (repo *UsersRepo) Insert(ctx context.Context, user *entity.User) (*entity.User, error) {
	queryable, err := GetQueryable(ctx, pgxPoolAdapter{repo.Pool})
	if err != nil {
		return nil, err
	}
	query, args, err := repo.Builder.Insert(
		`"user"(username, password, email, first_name, last_name, photo_url, birth_date, about_me)`,
	).
		Values(user.Username, user.Password, user.Email, user.FirstName, user.LastName, user.PhotoUrl, user.BirthDate, user.AboutMe).
		Suffix("RETURNING *").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %w", err)
	}
	rows, err := queryable.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}
	res, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.User])
	if err != nil {
		if getPgErrCode(err) == UniqueViolationErrCode {
			return nil, storage.ErrAlreadyExists
		}
		return nil, fmt.Errorf("failed to collect row into user struct: %w", err)
	}
	return res[0], nil
}
func (repo *UsersRepo) CheckExistsWithEmail(ctx context.Context, email string) (bool, error) {
	queryable, err := GetQueryable(ctx, pgxPoolAdapter{repo.Pool})
	if err != nil {
		return false, err
	}
	query, args, err := repo.Builder.Select("id").From(`"user"`).Where(squirrel.Eq{"email": email}).ToSql()
	if err != nil {
		return false, fmt.Errorf("failed to build sql query: %w", err)
	}
	row := queryable.QueryRow(ctx, query, args...)
	var userId int
	if err := row.Scan(&userId); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
func (repo *UsersRepo) GetByLogin(ctx context.Context, login string) (*entity.User, error) {
	queryable, err := GetQueryable(ctx, pgxPoolAdapter{repo.Pool})
	if err != nil {
		return nil, err
	}
	query, args, err := repo.Builder.Select("*").From(`"user"`).
		Where(squirrel.Or{squirrel.Eq{"email": login}, squirrel.Eq{"username": login}}).ToSql()
	fmt.Println("query", query, args)
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %w", err)
	}
	rows, err := queryable.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}
	res, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[entity.User])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row into user struct: %w", err)
	}
	if len(res) == 0 {
		return nil, storage.ErrNotFound
	}
	return res[0], nil
}
func (repo *UsersRepo) GetByID(ctx context.Context, id int) (*entity.User, error) {
	return nil, nil
}
func (repo *UsersRepo) UpdateIsActiveById(ctx context.Context, userId int, isActive bool) (*entity.User, error) {
	return nil, nil
}
func (repo *UsersRepo) AddPasskeyCredential(ctx context.Context, userId int, cred *entity.PasskeyCredential) error {
	return nil
}

package users_repo

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

type Repository struct {
	*postgres.Postgres
}

func (repo *Repository) Insert(ctx context.Context, user *entity.User) (*entity.User, error) {
	query := repo.Builder.Insert(
		`"user"(username, password, email, first_name, last_name, photo_url, birth_date, about_me, is_active)`,
	).
		Values(user.Username, user.Password, user.Email, user.FirstName, user.LastName, user.PhotoUrl, user.BirthDate, user.AboutMe, user.IsActive).
		Suffix("RETURNING *")
	insertedUser, err := postgres.ExecAndGetOne[entity.User](ctx, query, repo.Pool, nil)
	if err != nil {
		if postgres.IsUniqueViolationError(err) {
			return nil, storage.ErrAlreadyExists
		}
		return nil, err
	}
	if user.TwoFactorAuth != nil {
		query = repo.Builder.Insert(
			`"two_factor_auth"(user_id, enabled, transport, contact, totp_secret)`,
		).
			Values(insertedUser.ID, user.TwoFactorAuth.Enabled, user.TwoFactorAuth.Transport, user.TwoFactorAuth.Contact, user.TwoFactorAuth.TotpSecret).
			Suffix("RETURNING *")
		twoFA, err := postgres.ExecAndGetOne[entity.TwoFactorAuth](ctx, query, repo.Pool, nil)
		if err != nil {
			return nil, err
		}
		insertedUser.TwoFactorAuth = twoFA
	}
	return insertedUser, nil
}

func (repo *Repository) CheckExistsWithEmail(ctx context.Context, email string) (bool, error) {
	queryable, err := postgres.GetQueryable(ctx, postgres.PgxPoolAdapter{repo.Pool})
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

func (repo *Repository) GetByLogin(ctx context.Context, login string) (*entity.User, error) {
	qb := repo.Builder.Select(userSelect).From(`"user"`).
		LeftJoin(`two_factor_auth ON two_factor_auth.user_id="user".id`).
		Where(squirrel.Or{squirrel.Eq{"email": login}, squirrel.Eq{"username": login}})
	return postgres.ExecAndGetOne[entity.User](ctx, qb, repo.Pool, parseUserFromRow)
}

func (repo *Repository) fetchPasskeyCredentials(ctx context.Context, userId int) ([]entity.PasskeyCredential, error) {
	query := repo.Builder.Select("*").From("passkey_credential").Where(squirrel.Eq{"user_id": userId})
	creds, err := postgres.ExecAndGetMany[entity.PasskeyCredential](ctx, query, repo.Pool, nil)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, err
	}
	return creds, nil
}

func (repo *Repository) GetByID(ctx context.Context, id int) (*entity.User, error) {
	query := repo.Builder.Select(userSelect).From(`"user"`).
		LeftJoin(`two_factor_auth ON two_factor_auth.user_id="user".id`).
		Where(squirrel.Eq{"id": id})
	return postgres.ExecAndGetOne(ctx, query, repo.Pool, parseUserFromRow)
}

func (repo *Repository) UpdateIsActiveById(ctx context.Context, userId int, isActive bool) (*entity.User, error) {
	query := repo.Builder.Update(`"user"`).Set("is_active", isActive).
		Where(squirrel.Eq{"id": userId}).Suffix("RETURNING *")
	return postgres.ExecAndGetOne[entity.User](ctx, query, repo.Pool, nil)
}

func (repo *Repository) AddPasskeyCredential(ctx context.Context, userId int, cred *entity.PasskeyCredential) error {
	qb := repo.Builder.Insert(`"passkey_credential"(id, public_key, user_id, transports, backed_up)`).
		Values(cred.ID, cred.PublicKey, userId, cred.Transports, cred.BackedUp)
	queryable, err := postgres.GetQueryable(ctx, postgres.PgxPoolAdapter{repo.Pool})
	if err != nil {
		return err
	}
	query, args, err := qb.ToSql()
	if err != nil {
		return err
	}
	if _, err := queryable.Exec(ctx, query, args...); err != nil {
		if postgres.IsForeightKeyViolationError(err) {
			return storage.ErrNotFound
		}
		return err
	}
	return nil
}

func (repo *Repository) GetByIDWithPasskeyCredentials(ctx context.Context, userId int) (*entity.User, error) {
	user, err := repo.GetByID(ctx, userId)
	if err != nil {
		return nil, err
	}
	creds, err := repo.fetchPasskeyCredentials(ctx, userId)
	if err != nil {
		return nil, err
	}
	user.PasskeyCredentials = creds
	return user, nil
}

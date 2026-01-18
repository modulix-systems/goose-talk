package pgrepos

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage/pgrepos/sqlutils"
	"github.com/modulix-systems/goose-talk/postgres"
)

type UsersRepo struct {
	*postgres.Postgres
}

func (repo *UsersRepo) Save(ctx context.Context, user *entity.User) (*entity.User, error) {
	query := repo.Builder.Insert(`"user"`).
		Columns("username", "password", "email", "first_name", "last_name", "photo_url", "birth_date", "about_me", "is_active", "private_key").
		Values(user.Username, user.Password, user.Email, user.FirstName, user.LastName, user.PhotoUrl, user.BirthDate, user.AboutMe, user.IsActive, user.PrivateKey).
		Suffix("RETURNING *")
	savedUser, err := postgres.ExecAndGetOne[entity.User](ctx, query, repo.Pool, nil, repo.TransactionCtxKey)
	if err != nil {
		if errors.Is(err, postgres.ErrUniqueViolation) {
			return nil, storage.ErrAlreadyExists
		}
		return nil, err
	}
	if user.TwoFactorAuth != nil {
		user.TwoFactorAuth.UserId = savedUser.Id
		twoFA, err := repo.CreateTwoFa(ctx, user.TwoFactorAuth)
		if err != nil {
			return nil, err
		}
		savedUser.TwoFactorAuth = twoFA
	}
	return savedUser, nil
}

func (repo *UsersRepo) CheckExistsWithEmail(ctx context.Context, email string) (bool, error) {
	queryable, err := postgres.GetQueryable(ctx, repo.Pool, repo.TransactionCtxKey)
	if err != nil {
		return false, err
	}
	if r, ok := queryable.(postgres.Releaseable); ok {
		defer r.Release()
	}

	query, args := repo.Builder.Select("id").From(`"user"`).Where(squirrel.Eq{"email": email}).MustSql()

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
	qb := repo.Builder.Select(sqlutils.UserSelect).From(`"user"`).
		LeftJoin(`two_factor_auth ON two_factor_auth.user_id="user".id`).
		Where(squirrel.Or{squirrel.Eq{"email": login}, squirrel.Eq{"username": login}})
	user, err := postgres.ExecAndGetOne(ctx, qb, repo.Pool, sqlutils.RowToUser, repo.TransactionCtxKey)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (repo *UsersRepo) fetchPasskeyCredentials(ctx context.Context, userId int) ([]entity.PasskeyCredential, error) {
	query := repo.Builder.Select("*").From("passkey_credential").Where(squirrel.Eq{"user_id": userId})
	creds, err := postgres.ExecAndGetMany[entity.PasskeyCredential](ctx, query, repo.Pool, nil, repo.TransactionCtxKey)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, err
	}
	return creds, nil
}

func (repo *UsersRepo) GetByID(ctx context.Context, id int) (*entity.User, error) {
	query := repo.Builder.Select(sqlutils.UserSelect).From(`"user"`).
		LeftJoin(`two_factor_auth ON two_factor_auth.user_id="user".id`).
		Where(squirrel.Eq{"id": id})
	user, err := postgres.ExecAndGetOne(ctx, query, repo.Pool, sqlutils.RowToUser, repo.TransactionCtxKey)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (repo *UsersRepo) UpdateIsActiveById(ctx context.Context, userId int, isActive bool) (*entity.User, error) {
	query := repo.Builder.Update(`"user"`).Set("is_active", isActive).
		Where(squirrel.Eq{"id": userId}).Suffix("RETURNING *")
	user, err := postgres.ExecAndGetOne[entity.User](ctx, query, repo.Pool, nil, repo.TransactionCtxKey)
	if err != nil {
		if errors.Is(err, postgres.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (repo *UsersRepo) CreatePasskeyCredential(ctx context.Context, userId int, cred *entity.PasskeyCredential) error {
	qb := repo.Builder.Insert(`"passkey_credential"`).
		Columns("id", "public_key", "user_id", "transports", "backed_up").
		Values(cred.ID, cred.PublicKey, userId, cred.Transports, cred.BackedUp)
	if _, err := postgres.Exec(ctx, qb, repo.Pool, repo.TransactionCtxKey); err != nil {
		if errors.Is(err, postgres.ErrForeignKeyViolation) {
			return storage.ErrNotFound
		}
		return err
	}

	return nil
}

func (repo *UsersRepo) GetByIDWithPasskeyCredentials(ctx context.Context, userId int) (*entity.User, error) {
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

func (repo *UsersRepo) CreateTwoFa(ctx context.Context, ent *entity.TwoFactorAuth) (*entity.TwoFactorAuth, error) {
	qb := repo.Builder.Insert("two_factor_auth").
		Columns("user_id", "transport", "contact", "totp_secret").
		Values(ent.UserId, ent.Method, ent.Contact, ent.TotpSecret).
		Suffix("RETURNING *")

	twoFA, err := postgres.ExecAndGetOne[entity.TwoFactorAuth](ctx, qb, repo.Pool, nil, repo.TransactionCtxKey)

	if err != nil {
		if errors.Is(err, postgres.ErrForeignKeyViolation) {
			return nil, storage.ErrNotFound
		}
		if errors.Is(err, postgres.ErrUniqueViolation) {
			return nil, storage.ErrAlreadyExists
		}
		return nil, err
	}

	return twoFA, nil
}

func (repo *UsersRepo) UpdateTwoFaContact(ctx context.Context, userId int, contact string) error {
	qb := repo.Builder.Update("two_factor_auth").Set("contact", contact).Where(squirrel.Eq{"user_id": userId})
	if _, err := postgres.Exec(ctx, qb, repo.Pool, repo.TransactionCtxKey); err != nil {
		if errors.Is(err, postgres.ErrForeignKeyViolation) {
			return storage.ErrNotFound
		}
		return err
	}

	return nil
}

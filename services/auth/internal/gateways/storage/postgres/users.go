package postgres_repos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/utils"
	"github.com/modulix-systems/goose-talk/pkg/postgres"
)

const userColPrefix = "user"
const twoFaColPrefix = "two_factor_auth"

func formatColName(prefix string, name string) string {
	colName := utils.ToSnakeCase(name)
	if prefix != "" {
		return fmt.Sprintf("%s.%s", prefix, colName)
	}
	return colName
}

func getSelectColNameWithAlias(tableName string, name string) string {
	return fmt.Sprintf(
		`%s as "%s"`, formatColName(fmt.Sprintf(`"%s"`, tableName), name),
		formatColName(strings.ReplaceAll(tableName, "\"", ""), name),
	)
}

func getUserSelectCols() []string {
	userType := reflect.TypeOf(entity.User{})
	colsList := make([]string, 0, userType.NumField()+1)
	for i := 0; i < userType.NumField(); i++ {
		field := userType.Field(i)
		fieldType := field.Type
		if !utils.IsScalarType(fieldType) {
			if fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
			}
			switch fieldType.Name() {
			case "TwoFactorAuth":
				for i := 0; i < fieldType.NumField(); i++ {
					colsList = append(colsList, getSelectColNameWithAlias(twoFaColPrefix, fieldType.Field(i).Name))
				}
				continue
			case "Time":
			default:
				continue
			}
		}
		colsList = append(colsList, getSelectColNameWithAlias(userColPrefix, field.Name))
	}
	return colsList
}

var userSelect = strings.Join(getUserSelectCols(), ",")

func RowToUserStruct(row pgx.CollectableRow) (entity.User, error) {
	var user entity.User
	userData, err := pgx.RowToMap(row)
	if err != nil {
		return user, err
	}
	// group related entities flat data into map
	userStructJsonData := make(map[string]any)
	for key, value := range userData {
		keyParts := strings.Split(key, ".")
		var prefix, fieldName string = keyParts[0], keyParts[1]
		switch prefix {
		case userColPrefix:
			userStructJsonData[fieldName] = value
		case twoFaColPrefix:
			twoFaData, ok := userStructJsonData[twoFaColPrefix].(map[string]any)
			if !ok {
				twoFaData = make(map[string]any)
				userStructJsonData[twoFaColPrefix] = twoFaData
			}
			twoFaData[fieldName] = value
		}
	}
	userJson, err := json.Marshal(userStructJsonData)
	if err != nil {
		return user, err
	}
	if err := json.Unmarshal(userJson, &user); err != nil {
		return user, err
	}
	return user, nil
}

type UsersRepo struct {
	*postgres.Postgres
}

func (repo *UsersRepo) Insert(ctx context.Context, user *entity.User) (*entity.User, error) {
	query := repo.Builder.Insert(
		`"user"(username, password, email, first_name, last_name, photo_url, birth_date, about_me, is_active)`,
	).
		Values(user.Username, user.Password, user.Email, user.FirstName, user.LastName, user.PhotoUrl, user.BirthDate, user.AboutMe, user.IsActive).
		Suffix("RETURNING *")
	insertedUser, err := ExecAndGetOne[entity.User](ctx, query, repo.Pool, nil)
	if err != nil {
		if getPgErrCode(err) == UniqueViolationErrCode {
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
		twoFA, err := ExecAndGetOne[entity.TwoFactorAuth](ctx, query, repo.Pool, nil)
		if err != nil {
			return nil, err
		}
		insertedUser.TwoFactorAuth = twoFA
	}
	return insertedUser, nil
}
func (repo *UsersRepo) CheckExistsWithEmail(ctx context.Context, email string) (bool, error) {
	queryable, err := GetQueryable(ctx, PgxPoolAdapter{repo.Pool})
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
	qb := repo.Builder.Select(userSelect).From(`"user"`).
		LeftJoin(`two_factor_auth ON two_factor_auth.user_id="user".id`).
		Where(squirrel.Or{squirrel.Eq{"email": login}, squirrel.Eq{"username": login}})
	return ExecAndGetOne[entity.User](ctx, qb, repo.Pool, RowToUserStruct)
}

func (repo *UsersRepo) fetchPasskeyCredentials(ctx context.Context, userId int) ([]entity.PasskeyCredential, error) {
	query := repo.Builder.Select("*").From("passkey_credential").Where(squirrel.Eq{"user_id": userId})
	creds, err := ExecAndGetMany[entity.PasskeyCredential](ctx, query, repo.Pool, nil)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, err
	}
	return creds, nil
}

func (repo *UsersRepo) GetByID(ctx context.Context, id int) (*entity.User, error) {
	query := repo.Builder.Select(userSelect).From(`"user"`).
		LeftJoin(`two_factor_auth ON two_factor_auth.user_id="user".id`).
		Where(squirrel.Eq{"id": id})
	return ExecAndGetOne[entity.User](ctx, query, repo.Pool, RowToUserStruct)
}

func (repo *UsersRepo) UpdateIsActiveById(ctx context.Context, userId int, isActive bool) (*entity.User, error) {
	query := repo.Builder.Update(`"user"`).Set("is_active", isActive).
		Where(squirrel.Eq{"id": userId}).Suffix("RETURNING *")
	return ExecAndGetOne[entity.User](ctx, query, repo.Pool, nil)
}
func (repo *UsersRepo) AddPasskeyCredential(ctx context.Context, userId int, cred *entity.PasskeyCredential) error {
	qb := repo.Builder.Insert(`"passkey_credential"(id, public_key, user_id, transports, backed_up)`).
		Values(cred.ID, cred.PublicKey, userId, cred.Transports, cred.BackedUp)
	queryable, err := GetQueryable(ctx, PgxPoolAdapter{repo.Pool})
	if err != nil {
		return err
	}
	query, args, err := qb.ToSql()
	if err != nil {
		return err
	}
	if _, err := queryable.Exec(ctx, query, args...); err != nil {
		if getPgErrCode(err) == ForeignKeyViolationErrCode {
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

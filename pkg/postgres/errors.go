package postgres

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

func getErrorCode(err error) string {
	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) {
		return pgxErr.Code
	}
	return ""
}

func IsUniqueViolationError(err error) bool {
	return getErrorCode(err) == "23505"
}

func IsForeightKeyViolationError(err error) bool {
	return getErrorCode(err) == "23503"
}

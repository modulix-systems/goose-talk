package postgres

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNoRows              = errors.New("no rows found")
	ErrUniqueViolation     = errors.New("unique violation")
	ErrForeignKeyViolation = errors.New("foreign key violation")
)

func MapPgxError(err error) error {
	var pgxErr *pgconn.PgError
	wrap := func(mappedErr error) error {
		return fmt.Errorf("%w: %w", mappedErr, err)
	}
	if errors.As(err, &pgxErr) {
		switch pgxErr.Code {
		case "23505":
			return wrap(ErrUniqueViolation)
		case "23503":
			return wrap(ErrForeignKeyViolation)
		default:
			return err
		}
	}

	return err
}

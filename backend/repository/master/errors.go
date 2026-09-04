package master

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrConcurrentUpdate = errors.New("resource was changed by another request")

func IsUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == "23505"
}

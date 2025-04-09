package helpers

import "github.com/jackc/pgx/v5/pgconn"

func IsErrorCode(err error, code string) bool {
	if pgError, ok := err.(*pgconn.PgError); ok {
		return pgError.Code == code
	}
	return false
}

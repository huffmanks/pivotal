package db

import (
	"errors"

	"github.com/ncruces/go-sqlite3"
)

func IsUniqueConstraintError(err error) bool {
	var sqliteErr *sqlite3.Error

	if !errors.As(err, &sqliteErr) {
		return false
	}

	return sqliteErr.Code() == sqlite3.CONSTRAINT &&
		sqliteErr.ExtendedCode() == sqlite3.CONSTRAINT_UNIQUE
}

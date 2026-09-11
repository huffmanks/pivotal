package db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
)

func Scan[T any](scanner interface {
	Scan(dest ...any) error
}) (T, error) {
	var result T
	value := reflect.ValueOf(&result).Elem()

	if value.Kind() != reflect.Struct {
		return result, fmt.Errorf("db.Scan requires a struct type")
	}

	dest, err := scanDestinations(value)
	if err != nil {
		return result, err
	}

	if err := scanner.Scan(dest...); err != nil {
		return result, err
	}

	return result, nil
}

func scanDestinations(value reflect.Value) ([]any, error) {
	typeInfo := value.Type()
	dest := make([]any, 0, value.NumField())

	for i := 0; i < value.NumField(); i++ {
		fieldInfo := typeInfo.Field(i)

		if fieldInfo.PkgPath != "" || fieldInfo.Tag.Get("db") == "-" {
			continue
		}

		field := value.Field(i)
		if !field.CanAddr() {
			return nil, fmt.Errorf("db.Scan cannot address field %q", fieldInfo.Name)
		}

		dest = append(dest, field.Addr().Interface())
	}

	return dest, nil
}

func List[T any](
	ctx context.Context,
	db *sql.DB,
	query string,
	args ...any,
) ([]T, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]T, 0)

	for rows.Next() {
		item, err := Scan[T](rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func GetByID[T any](
	ctx context.Context,
	db *sql.DB,
	query string,
	id int64,
) (T, bool, error) {
	var zero T

	item, err := Scan[T](
		db.QueryRowContext(ctx, query, id),
	)

	if err == sql.ErrNoRows {
		return zero, false, nil
	}

	if err != nil {
		return zero, false, err
	}

	return item, true, nil
}

package db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

func Scan[T any](scanner interface {
	Scan(dest ...any) error
}) (T, error) {
	var result T

	value := reflect.ValueOf(&result).Elem()
	if value.Kind() != reflect.Struct {
		return result, fmt.Errorf("db.Scan requires a struct type")
	}

	typeInfo := value.Type()

	dest := make([]any, 0, value.NumField())

	for i := 0; i < value.NumField(); i++ {
		fieldInfo := typeInfo.Field(i)

		if fieldInfo.PkgPath != "" {
			continue
		}

		if fieldInfo.Tag.Get("db") == "-" {
			continue
		}

		field := value.Field(i)

		if !field.CanAddr() {
			return result, fmt.Errorf("db.Scan cannot address field %q", fieldInfo.Name)
		}

		dest = append(dest, field.Addr().Interface())
	}

	if err := scanner.Scan(dest...); err != nil {
		return result, err
	}

	return result, nil
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

func Insert(
	tx *sql.Tx,
	table string,
	value any,
) error {
	fields, values, err := structFields(value)
	if err != nil {
		return err
	}

	placeholders := make([]string, len(fields))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err = tx.Exec(query, values...)
	return err
}

func structFields(value any) ([]string, []any, error) {
	v := reflect.ValueOf(value)

	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, nil, fmt.Errorf("db.Insert requires a non-nil struct")
		}

		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("db.Insert requires a struct value")
	}

	t := v.Type()

	fields := make([]string, 0, v.NumField())
	values := make([]any, 0, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)

		if field.PkgPath != "" {
			continue
		}

		column := field.Tag.Get("db")
		if column == "-" {
			continue
		}

		if column == "" {
			column = field.Name
		}

		fields = append(fields, column)
		values = append(values, v.Field(i).Interface())
	}

	return fields, values, nil
}

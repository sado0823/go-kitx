package sqlx

import (
	"context"
	"database/sql"
)

type (
	transactionI interface {
		Session
		Commit() error
		Rollback() error
	}

	transaction struct {
		*sql.Tx
	}
)

var begin = func(db *sql.DB) (transactionI, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	return &transaction{tx}, nil
}

func (c *transaction) Exec(ctx context.Context, query string, args ...interface{}) (result sql.Result, err error) {
	startCtx, span := startSpan(ctx, "Transaction Exec")
	defer func() { endSpan(span, err) }()

	result, err = c.ExecContext(startCtx, query, args...)
	return result, err
}

func (c *transaction) QueryRow(ctx context.Context, v interface{}, query string, args ...interface{}) (err error) {
	startCtx, span := startSpan(ctx, "Transaction QueryRow")
	defer func() { endSpan(span, err) }()

	return c.query(startCtx, func(rows *sql.Rows) error {
		return unmarshalRow(v, rows, true)
	}, query, args...)
}

func (c *transaction) Query(ctx context.Context, v interface{}, query string, args ...interface{}) (err error) {
	startCtx, span := startSpan(ctx, "Transaction Query")
	defer func() { endSpan(span, err) }()

	return c.query(startCtx, func(rows *sql.Rows) error {
		return unmarshalRows(v, rows, true)
	}, query, args...)
}

func (c *transaction) query(ctx context.Context, scanner func(rows *sql.Rows) error, query string, args ...interface{}) (err error) {
	rows, err := c.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	return scanner(rows)
}

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

func (c *transaction) Prepare(ctx context.Context, query string) (stmt StmtSession, err error) {
	startCtx, span := startSpan(ctx, "Transaction Prepare")
	defer func() { endSpan(span, err) }()

	var sqlStmt *sql.Stmt
	sqlStmt, err = c.PrepareContext(startCtx, query)
	if err != nil {
		return nil, err
	}

	return &statement{stmt: sqlStmt, query: query}, nil
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
	if err = scanner(rows); err != nil {
		return err
	}

	return rows.Err()
}

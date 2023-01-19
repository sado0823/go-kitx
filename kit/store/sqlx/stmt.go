package sqlx

import (
	"context"
	"database/sql"
)

type (
	StmtSession interface {
		Exec(ctx context.Context, args ...interface{}) (result sql.Result, err error)
		QueryRow(ctx context.Context, v interface{}, args ...interface{}) error
		Query(ctx context.Context, v interface{}, args ...interface{}) error
		Close() error
	}

	statement struct {
		query string
		stmt  *sql.Stmt
	}
)

func (s *statement) Close() error {
	return s.stmt.Close()
}

func (s *statement) Exec(ctx context.Context, args ...interface{}) (result sql.Result, err error) {
	startCtx, span := startSpan(ctx, "Prepare Exec")
	defer func() { endSpan(span, err) }()

	return s.stmt.ExecContext(startCtx, args...)
}

func (s *statement) QueryRow(ctx context.Context, v interface{}, args ...interface{}) (err error) {
	startCtx, span := startSpan(ctx, "Prepare QueryRow")
	defer func() { endSpan(span, err) }()

	return s.doQuery(startCtx, func(rows *sql.Rows) error {
		return unmarshalRow(v, rows, true)
	}, args...)
}

func (s *statement) Query(ctx context.Context, v interface{}, args ...interface{}) (err error) {
	startCtx, span := startSpan(ctx, "Prepare Query")
	defer func() { endSpan(span, err) }()

	return s.doQuery(startCtx, func(rows *sql.Rows) error {
		return unmarshalRows(v, rows, true)
	}, args...)
}

func (s *statement) doQuery(ctx context.Context, scanner func(*sql.Rows) error, args ...interface{}) error {

	rows, err := s.stmt.QueryContext(ctx, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if err = scanner(rows); err != nil {
		return err
	}

	return rows.Err()
}

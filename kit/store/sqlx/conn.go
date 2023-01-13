package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/sado0823/go-kitx/kit/breaker"
)

var (
	ErrNotFound                 = sql.ErrNoRows
	ErrUnsupportedUnmarshalType = fmt.Errorf("unsupported unmarshal type")
	ErrNotSettable              = fmt.Errorf("not a settable type")
	ErrColumnsNotMatched        = fmt.Errorf("columns not matched")
	ErrNotReadable              = fmt.Errorf("not a readable type")
)

func New(driveName, datasource string) (Conn, error) {
	db, err := open(driveName, datasource)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &conn{
		db:  db,
		tx:  begin,
		brk: breaker.New(),
	}, nil
}

func NewWith(db *sql.DB) (Conn, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &conn{
		db:  db,
		tx:  begin,
		brk: breaker.New(),
	}, nil
}

func open(driveName, datasource string) (*sql.DB, error) {
	conn, err := sql.Open(driveName, datasource)
	if err != nil {
		return nil, err
	}

	conn.SetMaxIdleConns(64)
	conn.SetMaxOpenConns(64)
	conn.SetConnMaxLifetime(time.Minute)

	return conn, nil
}

type (
	Session interface {
		Exec(ctx context.Context, query string, args ...interface{}) (result sql.Result, err error)
		QueryRow(ctx context.Context, v interface{}, query string, args ...interface{}) error
		Query(ctx context.Context, v interface{}, query string, args ...interface{}) error
	}

	Conn interface {
		Session
		Transaction(ctx context.Context, fn func(ctx context.Context, session Session) error) error
		Close() error
	}
)

type conn struct {
	db  *sql.DB
	tx  func(db *sql.DB) (transactionI, error)
	brk breaker.Breaker
}

func acceptable(err error) bool {
	return err == nil || err == sql.ErrNoRows || err == sql.ErrTxDone
}

func (c *conn) Close() error {
	return c.db.Close()
}

func (c *conn) Transaction(ctx context.Context, fn func(ctx context.Context, session Session) error) (err error) {
	startCtx, span := startSpan(ctx, "Transaction")
	defer func() { endSpan(span, err) }()

	return c.brk.DoWithAcceptable(func() (err error) {
		var tx transactionI
		tx, err = c.tx(c.db)
		if err != nil {
			return err
		}

		defer func() {
			if re := recover(); re != nil {
				if e := tx.Rollback(); e != nil {
					err = fmt.Errorf("recover from:%#v, tx rollback failed:%w", re, e)
				} else {
					err = fmt.Errorf("recover from:%#v", re)
				}
			} else if err != nil {

			} else {
				err = tx.Commit()
			}
		}()

		return fn(startCtx, tx)
	}, acceptable)
}

func (c *conn) Exec(ctx context.Context, query string, args ...interface{}) (result sql.Result, err error) {
	startCtx, span := startSpan(ctx, "Exec")
	defer func() { endSpan(span, err) }()

	err = c.brk.DoWithAcceptable(func() error {
		result, err = c.db.ExecContext(startCtx, query, args...)
		return err
	}, acceptable)

	return result, err
}

func (c *conn) QueryRow(ctx context.Context, v interface{}, query string, args ...interface{}) (err error) {
	startCtx, span := startSpan(ctx, "QueryRow")
	defer func() { endSpan(span, err) }()

	return c.query(startCtx, func(rows *sql.Rows) error {
		return unmarshalRow(v, rows, true)
	}, query, args...)
}

func (c *conn) Query(ctx context.Context, v interface{}, query string, args ...interface{}) (err error) {
	startCtx, span := startSpan(ctx, "Query")
	defer func() { endSpan(span, err) }()

	return c.query(startCtx, func(rows *sql.Rows) error {
		return unmarshalRows(v, rows, true)
	}, query, args...)
}

func (c *conn) query(ctx context.Context, scanner func(rows *sql.Rows) error, query string, args ...interface{}) (err error) {
	return c.brk.DoWithAcceptable(func() error {
		rows, err := c.db.QueryContext(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		if err = scanner(rows); err != nil {
			return err
		}
		return rows.Err()
	}, acceptable)
}

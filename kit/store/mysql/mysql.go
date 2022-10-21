package mysql

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Mysql struct {
	*sql.DB
}

func New(datasource string) (*Mysql, error) {
	conn, err := sql.Open("mysql", datasource)
	if err != nil {
		return nil, err
	}

	// we need to do this until the issue https://github.com/golang/go/issues/9851 get fixed
	// discussed here https://github.com/go-sql-driver/mysql/issues/257
	// if the discussed SetMaxIdleTimeout methods added, we'll change this behavior
	// 8 means we can't have more than 8 goroutines to concurrently access the same database.
	conn.SetMaxIdleConns(64)
	conn.SetMaxOpenConns(64)
	conn.SetConnMaxLifetime(time.Minute)

	return &Mysql{conn}, nil
}

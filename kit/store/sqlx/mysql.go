package sqlx

import (
	_ "github.com/go-sql-driver/mysql"

	"github.com/sado0823/go-kitx/kit/breaker"
)

func NewMysql(datasource string) (Conn, error) {
	db, err := open("mysql", datasource)
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

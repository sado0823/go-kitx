package sqlx

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestQueryWithError(t *testing.T) {
	type city struct {
		ID    int64  `json:"id" db:"id"`
		Name  string `json:"name" db:"name"`
		State int64  `json:"state" db:"state"`
		Ctime string `json:"ctime" db:"ctime"`
		Mtime string `json:"mtime" db:"mtime"`
	}

	needErr := fmt.Errorf("rows query with breaker error")

	runSqlMockTest(t, func(ctx context.Context, conn Conn, mock sqlmock.Sqlmock) {
		mock.ExpectQuery("select (.*) from `test`").
			WillReturnError(needErr)

		var records []*city
		err := conn.Query(ctx, &records, "select * from `test`")
		assert.EqualError(t, err, needErr.Error())
	})
}

func TestQueryRows(t *testing.T) {
	type city struct {
		ID    int64  `json:"id" db:"id"`
		Name  string `json:"name" db:"name"`
		State int64  `json:"state" db:"state"`
		Ctime string `json:"ctime" db:"ctime"`
		Mtime string `json:"mtime" db:"mtime"`
	}

	runSqlMockTest(t, func(ctx context.Context, conn Conn, mock sqlmock.Sqlmock) {
		rows := mock.NewRows([]string{"id", "name", "state", "ctime", "mtime"}).
			AddRow(2, "bar", 2, "2021-01-02 12:11:10", "2022-01-01 12:11:10").
			AddRow(1, "foo", 1, "2021-01-01 12:11:10", "2022-01-01 12:11:10")

		mock.ExpectQuery("select (.*) from `test` order by id desc").
			WillReturnRows(rows)

		var records []*city
		err := conn.Query(ctx, &records, "select * from `test` order by id desc")
		assert.Nil(t, err)
		assert.Equal(t, 2, len(records))
		for i, record := range records {
			t.Log(i, record)
		}
	})
}

func TestExecUpdate(t *testing.T) {
	runSqlMockTest(t, func(ctx context.Context, conn Conn, mock sqlmock.Sqlmock) {
		mock.ExpectExec("update test").
			WithArgs("foo", 123).WillReturnResult(sqlmock.NewResult(0, 1))

		result, err := conn.Exec(ctx, "update test set name = ? where name = ? and id = ?", "foo", 123)
		assert.Nil(t, err)
		affected, err := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(1), affected)
	})
}

func TestExecInsert(t *testing.T) {
	runSqlMockTest(t, func(ctx context.Context, conn Conn, mock sqlmock.Sqlmock) {
		mock.ExpectExec("insert into `test`").
			WithArgs("foo", 1, "bar", 2).WillReturnResult(sqlmock.NewResult(2, 2))

		result, err := conn.Exec(ctx, "insert into `test` values (?,?),(?,?)", "foo", 1, "bar", 2)
		assert.Nil(t, err)
		affected, err := result.RowsAffected()
		assert.Nil(t, err)
		assert.Equal(t, int64(2), affected)
	})
}

func TestQueryInt64(t *testing.T) {
	runSqlMockTest(t, func(ctx context.Context, conn Conn, mock sqlmock.Sqlmock) {
		rows := mock.NewRows([]string{"count"}).AddRow(2233)
		mock.ExpectQuery("select (.*) as count from `test` where id > ?").
			WithArgs(100).WillReturnRows(rows)

		var val int64
		err := conn.QueryRow(ctx, &val, "select count(*) as count from `test` where id > ?", 100)
		assert.Nil(t, err)
		assert.Equal(t, int64(2233), val)
		fmt.Println(val)
	})
}

func runSqlMockTest(t *testing.T, fn func(ctx context.Context, conn Conn, mock sqlmock.Sqlmock)) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("get sql mock failed, err:%+v", err)
	}
	with, err := NewWith(db)
	if err != nil {
		t.Fatalf("new with db err:%+v", err)
	}
	defer with.Close()

	fn(context.Background(), with, mock)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql mock not all expectations were met, err:%+v", err)
	}
}

//type city struct {
//	ID    int64  `json:"id" db:"id"`
//	Name  string `json:"name" db:"name"`
//	State int64  `json:"state" db:"state"`
//	Ctime string `json:"ctime" db:"ctime"`
//	Mtime string `json:"mtime" db:"mtime"`
//}

func Test_Mysql(t *testing.T) {
	//conn, err := NewMysql("root:admin@tcp(127.0.0.1:3306)/bilibili_aegis?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8mb4")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//records := make([]city, 0)
	//////records := make([]int64, 0)
	////err = conn.Query(context.Background(), &records, "select * from test222")
	////if err != nil {
	////	t.Fatal(err)
	////}
	////
	////for _, record := range records {
	////	t.Log(record)
	////}
	//
	////err = conn.Transaction(context.Background(), func(ctx context.Context, session Session) error {
	////	res, err := session.Exec(ctx, "insert into `test222` (name,state) values(?,?)", "tx2", -12)
	////	if err != nil {
	////		return err
	////	}
	////	id, err := res.LastInsertId()
	////	if err != nil {
	////		return err
	////	}
	////	t.Log("got last insert id: ", id)
	////	return fmt.Errorf("should tx rollback")
	////})
	////if err != nil {
	////	t.Fatal(err)
	////}
	//
	//prepare, err := conn.Prepare(context.Background(), "select * from test222 where id = ?")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//err = prepare.Query(context.Background(), &records, 3)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//for _, record := range records {
	//	t.Log(record)
	//}
}

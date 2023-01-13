package sqlx

import (
	"testing"
)

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
	////records := make([]int64, 0)
	//err = conn.Query(context.Background(), &records, "select * from test222")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//for _, record := range records {
	//	t.Log(record)
	//}
}

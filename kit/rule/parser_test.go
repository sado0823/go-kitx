package rule

import (
	"context"
	"testing"
)

func Test_New(t *testing.T) {
	// todo update ReadMe
	// support param struct Export filed
	// support param array get with index
	// support recursive param, get data with `.`
	type T struct {
		Name  string `json:"name"`
		Hobby []T    `json:"hobby"`
	}
	//expr := `(foo - 90 > 0 ) && ( foo > 1 || foo <1 ) && foo > 1`
	//expr := `foo.bar.Hobby.0.Name == "jay" && func test(foo.bar.Hobby.0,1,2,3)`
	//param := map[string]interface{}{
	//	"foo": map[string]interface{}{
	//		"bar": T{Name: "tom", Hobby: []T{{Name: "jay"}}},
	//	},
	//	"in": 12.2,
	//}
	tt := T{
		Name:  "ttt",
		Hobby: nil,
	}
	parser, err := New(context.Background(), "Name", WithCustomFn("test", func(evalParam interface{}, arguments ...interface{}) (interface{}, error) {
		t.Log("i am test func")
		t.Log("evalParam: ", evalParam)
		t.Log(arguments...)
		return true, nil
	}))
	if err != nil {
		panic(err)
	}
	res, err := parser.Eval(tt)
	if err != nil {
		panic(err)
	}
	t.Logf("res=%v\t type=%T\t err=%+v \n", res, res, err)
}

func Test_Do(t *testing.T) {
	//expr := `(foo - 90 > 0 ) && ( foo > 1 || foo <1 ) && foo > 1`
	expr := `func foo() + 1`
	param := map[string]interface{}{
		"foo": 5,
		"bar": 6,
	}

	res, err := Do(context.Background(), expr, param, WithCustomFn("foo", func(evalParam interface{}, arguments ...interface{}) (interface{}, error) {
		return 1, nil
	}))
	if err != nil {
		panic(err)
	}
	t.Logf("res=%v\t type=%T\t err=%+v \n", res, res, err)
}

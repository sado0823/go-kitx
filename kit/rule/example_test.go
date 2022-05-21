package rule_test

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/sado0823/go-kitx/kit/rule"
)

func ExampleDo() {

	params := map[string]interface{}{"foo": 1}

	value, err := rule.Do(context.Background(), `foo + 1`, params)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Print(value)

	// Output:
	// 2
}

func ExampleNew() {
	params := map[string]interface{}{"foo": 1}

	parser, err := rule.New(context.Background(), `foo + 1`)
	if err != nil {
		fmt.Println(err)
	}

	value, err := parser.Eval(params)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Print(value)

	// Output:
	// 2
}

func ExampleWithCustomFn() {
	type Child struct {
		Name   string
		Age    int
		IsVIP  bool
		Map    map[string]int
		Nested *Child
	}
	type User struct {
		Name     string
		Age      int
		IsVIP    bool
		Nil      interface{}
		Children []Child
	}

	params := &User{
		Name:  "foo",
		Age:   18,
		IsVIP: true,
		Nil:   nil,
		Children: []Child{
			{
				// 0
				Name: "child0", Age: 0, IsVIP: false, Map: map[string]int{"child0": 0}, Nested: &Child{Name: "child0-child"},
			},
			{
				// 1
				Name: "child1", Age: 1, IsVIP: true, Map: map[string]int{"child1": 1}, Nested: &Child{},
			},
		},
	}

	value, err := rule.Do(
		context.Background(),
		`Name == "foo" && 
(Name + "bar" == "foobar") && 
(Age == 17 || Age == 18) &&
(Age + 1 == 19) && 
func in(Name,2,"foo",1) && 
func strlen("abc") == 3 && 
func isVIP() && 
Children.1.Name == "child1" && 
Children.1.Map.child1 == 1 && 
Children.0.Nested.Name == "child0-child"`,
		params,
		/* custom func `strlen` return args[0]'s count with float64 type */
		rule.WithCustomFn("strlen", func(evalParam interface{}, arguments ...interface{}) (interface{}, error) {
			if len(arguments) == 0 {
				return 0, nil
			}
			return float64(utf8.RuneCount([]byte(arguments[0].(string)))), nil
		}),
		/*custom func `isVIP` return if evalParam.IsVIP is true with bool type*/
		rule.WithCustomFn("isVIP", func(evalParam interface{}, arguments ...interface{}) (interface{}, error) {
			userCurrent := evalParam.(*User)
			return userCurrent.IsVIP == true, nil
		}),
	)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Print(value)

	// Output:
	// true
}

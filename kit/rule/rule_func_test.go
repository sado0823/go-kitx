package rule

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func Test_Func(t *testing.T) {

	type ext struct {
		name string
		fn   CustomFn
	}

	Function := func(name string, fn CustomFn) *ext {
		return &ext{name: name, fn: fn}
	}

	var funcCases = []struct {
		name      string
		expr      string
		want      interface{}
		extension *ext
	}{
		{
			name: "Single function",
			expr: "func foo()",
			extension: Function("foo", func(arguments ...interface{}) (interface{}, error) {
				return true, nil
			}),

			want: true,
		},
		{
			name: "Func with argument",
			expr: "func passthrough(1)",
			extension: Function("passthrough", func(arguments ...interface{}) (interface{}, error) {
				return arguments[0], nil
			}),
			want: 1.0,
		},
		{
			name: "Func with arguments",
			expr: "func passthrough(1, 2)",
			extension: Function("passthrough", func(arguments ...interface{}) (interface{}, error) {
				return arguments[0].(float64) + arguments[1].(float64), nil
			}),
			want: 3.0,
		},
		{
			name: "Nested function with operatorPrecedence",
			expr: "func sum(1, func sum(2, 3), 2 + 2, 2 * 2)",
			extension: Function("sum", func(arguments ...interface{}) (interface{}, error) {
				sum := 0.0
				for _, v := range arguments {
					sum += v.(float64)
				}
				return sum, nil
			}),
			want: 14.0,
		},
		{
			name: "Empty function and modifier, compared",
			expr: "func numeric()-1 > 0",
			extension: Function("numeric", func(arguments ...interface{}) (interface{}, error) {
				return 2.0, nil
			}),
			want: true,
		},
		{
			name: "Empty function comparator",
			expr: "func numeric() > 0",
			extension: Function("numeric", func(arguments ...interface{}) (interface{}, error) {
				return 2.0, nil
			}),
			want: true,
		},
		{

			name: "Empty function logical operator",
			expr: "func success() && !false",
			extension: Function("success", func(arguments ...interface{}) (interface{}, error) {
				return true, nil
			}),
			want: true,
		},
		{
			name: "Empty function with prefix",
			expr: "-func ten()",
			extension: Function("ten", func(arguments ...interface{}) (interface{}, error) {
				return 10.0, nil
			}),
			want: -10.0,
		},
		{
			name: "Empty function as part of chain",
			expr: "10 - func numeric() - 2",
			extension: Function("numeric", func(arguments ...interface{}) (interface{}, error) {
				return 5.0, nil
			}),
			want: 3.0,
		},
		{
			name: "Enclosed empty function with modifier and comparator (#28)",
			expr: "(func ten() - 1) > 3",
			extension: Function("ten", func(arguments ...interface{}) (interface{}, error) {
				return 10.0, nil
			}),
			want: true,
		},
		{
			name: "Variadic",
			expr: `func sum(1,2,3,4)`,
			extension: Function("sum", func(arguments ...interface{}) (interface{}, error) {
				sum := 0.
				for _, a := range arguments {
					sum += a.(float64)
				}
				return sum, nil
			}),
			want: 10.0,
		},
	}

	for _, funcCase := range funcCases {
		t.Run(funcCase.name, func(t *testing.T) {
			ctx := context.Background()

			// do check
			if err := Check(ctx, funcCase.expr, WithCustomFn(funcCase.extension.name, funcCase.extension.fn)); err != nil {
				t.Fatalf("Check() Name=%s expre=%s expected err nil but got %+v", funcCase.name, funcCase.expr, err)
				return
			}

			// do in once
			got, err := Do(ctx, funcCase.expr, nil, WithCustomFn(funcCase.extension.name, funcCase.extension.fn))
			if err != nil {
				t.Fatalf("Do() Name=%s expre=%s expected err nil but got %+v", funcCase.name, funcCase.expr, err)
				return
			} else if !reflect.DeepEqual(got, funcCase.want) {
				t.Fatalf("Do() Name=%s expre=%s expected %v but got %v", funcCase.name, funcCase.expr, funcCase.want, got)
			} else {
				fmt.Println(funcCase.name, " ", funcCase.expr, "=", got)
			}

			// do with new and step eval
			parser, err := New(ctx, funcCase.expr, WithCustomFn(funcCase.extension.name, funcCase.extension.fn))
			if err != nil {
				t.Fatalf("New() Name=%s expre=%s expected err nil but got %+v", funcCase.name, funcCase.expr, err)
				return
			}

			got, err = parser.Eval(nil)
			if err != nil {
				t.Fatalf("Eval() Name=%s expre=%s expected err nil but got %+v", funcCase.name, funcCase.expr, err)
				return
			} else if !reflect.DeepEqual(got, funcCase.want) {
				t.Fatalf("Eval() Name=%s expre=%s expected %v but got %v", funcCase.name, funcCase.expr, funcCase.want, got)
			}
		})
	}
}

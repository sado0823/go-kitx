package rule

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

type (
	dummyParameter struct {
		String    string
		Int       int
		BoolFalse bool
		Nil       interface{}
		Nested    dummyNestedParameter
	}
	dummyNestedParameter struct {
		Funk  string
		Map   map[string]int
		Slice []int
	}
)

func Test_Params(t *testing.T) {

	var foo = dummyParameter{
		String:    "string!",
		Int:       101,
		BoolFalse: false,
		Nil:       nil,
		Nested: dummyNestedParameter{
			Funk:  "funkalicious",
			Map:   map[string]int{"a": 1, "b": 2, "c": 3},
			Slice: []int{1, 2, 3},
		},
	}

	var paramsCases = []struct {
		name   string
		expr   string
		want   interface{}
		params interface{}
	}{
		{
			name: "Single parameter modified by constant",
			expr: "foo + 2",
			params: map[string]interface{}{
				"foo": 2.0,
			},
			want: 4.0,
		},
		{

			name: "Single parameter modified by variable",
			expr: "foo * bar",
			params: map[string]interface{}{
				"foo": 5.0,
				"bar": 2.0,
			},
			want: 10.0,
		},
		{

			name: "Multiple multiplications of the same parameter",
			expr: "foo * foo * foo",
			params: map[string]interface{}{
				"foo": 10.0,
			},
			want: 1000.0,
		},
		{

			name: "Multiple additions of the same parameter",
			expr: "foo + foo + foo",
			params: map[string]interface{}{
				"foo": 10.0,
			},
			want: 30.0,
		},
		{

			name: "Parameter name sensitivity",
			expr: "foo + FoO + FOO",
			params: map[string]interface{}{
				"foo": 8.0,
				"FoO": 4.0,
				"FOO": 2.0,
			},
			want: 14.0,
		},
		{

			name:   "Sign prefix comparison against prefixed variable",
			expr:   "-1 < foo",
			params: map[string]interface{}{"foo": 8.0},
			want:   true,
		},
		{

			name:   "Fixed-point parameter",
			expr:   "foo > 1",
			params: map[string]interface{}{"foo": 2},
			want:   true,
		},
		{

			name: "Modifier after closing clause",
			expr: "(2 + 2) + 2 == 6",
			want: true,
		},
		{

			name: "Comparator after closing clause",
			expr: "(2 + 2) >= 4",
			want: true,
		},
		{

			name: "String concat with single string parameter",
			expr: `foo + "bar"`,
			params: map[string]interface{}{
				"foo": "baz"},
			want: "bazbar",
		},
		{

			name: "String concat with multiple string parameter",
			expr: "foo + bar",
			params: map[string]interface{}{
				"foo": "baz",
				"bar": "quux",
			},
			want: "bazquux",
		},
		{

			name: "String concat with float parameter",
			expr: "foo + bar",
			params: map[string]interface{}{
				"foo": "baz",
				"bar": 123.0,
			},
			want: "baz123",
		},
		{

			name:   "Mixed multiple string concat",
			expr:   `foo + 123 + "bar" + "true"`,
			params: map[string]interface{}{"foo": "baz"},
			want:   "baz123bartrue",
		},
		{

			name: "Integer width spectrum",
			expr: "uint8 + uint16 + uint32 + uint64 + int8 + int16 + int32 + int64",
			params: map[string]interface{}{
				"uint8":  uint8(0),
				"uint16": uint16(0),
				"uint32": uint32(0),
				"uint64": uint64(0),
				"int8":   int8(0),
				"int16":  int16(0),
				"int32":  int32(0),
				"int64":  int64(0),
			},
			want: 0.0,
		},
		{

			name: "Two-boolean logical operation (for issue #8)",
			expr: "(foo == true) || (bar == true)",
			params: map[string]interface{}{
				"foo": true,
				"bar": false,
			},
			want: true,
		},
		{

			name: "Two-variable integer logical operation (for issue #8)",
			expr: "foo > 10 && bar > 10",
			params: map[string]interface{}{
				"foo": 1,
				"bar": 11,
			},
			want: false,
		},
		{

			name: "String concat with single string parameter",
			expr: `foo + "bar"`,
			params: map[string]interface{}{
				"foo": "baz"},
			want: "bazbar",
		},
		{

			name: "String concat with multiple string parameter",
			expr: "foo + bar",
			params: map[string]interface{}{
				"foo": "baz",
				"bar": "quux",
			},
			want: "bazquux",
		},
		{
			name: "Integer width spectrum",
			expr: "uint8 + uint16 + uint32 + uint64 + int8 + int16 + int32 + int64",
			params: map[string]interface{}{
				"uint8":  uint8(0),
				"uint16": uint16(0),
				"uint32": uint32(0),
				"uint64": uint64(0),
				"int8":   int8(0),
				"int16":  int16(0),
				"int32":  int32(0),
				"int64":  int64(0),
			},
			want: 0.0,
		},
		{

			name:   "Multiple comparator/logical operators (#30)",
			expr:   "(foo >= 2887057408 && foo <= 2887122943) || (foo >= 168100864 && foo <= 168118271)",
			params: map[string]interface{}{"foo": 2887057409},
			want:   true,
		},
		{

			name:   "Multiple comparator/logical operators, opposite order (#30)",
			expr:   "(foo >= 168100864 && foo <= 168118271) || (foo >= 2887057408 && foo <= 2887122943)",
			params: map[string]interface{}{"foo": 2887057409},
			want:   true,
		},
		{

			name:   "Multiple comparator/logical operators, small value (#30)",
			expr:   "(foo >= 2887057408 && foo <= 2887122943) || (foo >= 168100864 && foo <= 168118271)",
			params: map[string]interface{}{"foo": 168100865},
			want:   true,
		},
		{

			name:   "Multiple comparator/logical operators, small value, opposite order (#30)",
			expr:   "(foo >= 168100864 && foo <= 168118271) || (foo >= 2887057408 && foo <= 2887122943)",
			params: map[string]interface{}{"foo": 168100865},
			want:   true,
		},
		{

			name:   "Incomparable array equality comparison",
			expr:   "arr == arr",
			params: map[string]interface{}{"arr": []int{0, 0, 0}},
			want:   true,
		},
		{

			name:   "Incomparable array not-equality comparison",
			expr:   "arr != arr",
			params: map[string]interface{}{"arr": []int{0, 0, 0}},
			want:   false,
		},
		{
			name: "complex64 number as parameter",
			expr: "complex64",
			params: map[string]interface{}{
				"complex64":  complex64(0),
				"complex128": complex128(0),
			},
			want: complex64(0),
		},
		{
			name: "complex128 number as parameter",
			expr: "complex128",
			params: map[string]interface{}{
				"complex64":  complex64(0),
				"complex128": complex128(0),
			},
			want: complex128(0),
		},
		{
			name:   "Simple parameter call",
			expr:   "foo.String",
			params: map[string]interface{}{"foo": foo},
			want:   foo.String,
		},
		{
			name:   "Simple parameter call from pointer",
			expr:   "fooptr.String",
			params: map[string]interface{}{"fooptr": &foo},
			want:   foo.String,
		},
		{
			name:   "Simple parameter call",
			expr:   `foo.String == "hi"`,
			params: map[string]interface{}{"foo": foo},
			want:   false,
		},
		{
			name:   "Simple parameter call with modifier",
			expr:   `foo.String + "hi"`,
			params: map[string]interface{}{"foo": foo},
			want:   foo.String + "hi",
		},
		{
			name:   "Simple parameter call with array",
			expr:   `foo.Nested.Slice.1`,
			params: map[string]interface{}{"foo": foo},
			want:   2.0,
		},
		{
			name:   "Simple parameter call with array modifier",
			expr:   `foo.Nested.Slice.1 + 1`,
			params: map[string]interface{}{"foo": foo},
			want:   3.0,
		},
		{
			name:   "Nested parameter call",
			expr:   "foo.Nested.Funk",
			params: map[string]interface{}{"foo": foo},
			want:   "funkalicious",
		},
		{
			name:   "Nested map call",
			expr:   `foo.Nested.Map.a`,
			params: map[string]interface{}{"foo": foo},
			want:   1.0,
		},
		{
			name:   "Parameter call with + modifier",
			expr:   "1 + foo.Int",
			params: map[string]interface{}{"foo": foo},
			want:   102.0,
		},
		{
			name:   "Parameter call with && operator",
			expr:   "true && foo.BoolFalse",
			params: map[string]interface{}{"foo": foo},
			want:   false,
		},
	}

	for _, paramCase := range paramsCases {
		t.Run(paramCase.name, func(t *testing.T) {
			ctx := context.Background()

			// do check
			if err := Check(ctx, paramCase.expr); err != nil {
				t.Fatalf("Check() Name=%s expre=%s expected err nil but got %+v", paramCase.name, paramCase.expr, err)
				return
			}

			// do in once
			got, err := Do(ctx, paramCase.expr, paramCase.params)
			if err != nil {
				t.Fatalf("Do() Name=%s expre=%s expected err nil but got %+v", paramCase.name, paramCase.expr, err)
				return
			} else if !reflect.DeepEqual(got, paramCase.want) {
				t.Fatalf("Do() Name=%s expre=%s expected %v but got %v", paramCase.name, paramCase.expr, paramCase.want, got)
			} else {
				fmt.Println(paramCase.name, " ", paramCase.expr, "=", got)
			}

			// do with new and step eval
			parser, err := New(ctx, paramCase.expr)
			if err != nil {
				t.Fatalf("New() Name=%s expre=%s expected err nil but got %+v", paramCase.name, paramCase.expr, err)
				return
			}

			got, err = parser.Eval(paramCase.params)
			if err != nil {
				t.Fatalf("Eval() Name=%s expre=%s expected err nil but got %+v", paramCase.name, paramCase.expr, err)
				return
			} else if !reflect.DeepEqual(got, paramCase.want) {
				t.Fatalf("Eval() Name=%s expre=%s expected %v but got %v", paramCase.name, paramCase.expr, paramCase.want, got)
			}
		})
	}
}

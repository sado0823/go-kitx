package rule

import (
	"context"
	"reflect"
	"testing"
)

func Test_No_Params(t *testing.T) {
	var noParamsCases = []struct {
		name string
		expr string
		want interface{}
	}{
		{
			name: "Number",
			expr: "100",
			want: 100.0,
		},
		{
			name: "Single PLUS",
			expr: "51 + 49",
			want: 100.0,
		},
		{
			name: "Single MINUS",
			expr: "100 - 51",
			want: 49.0,
		},
		{
			name: "Single BITWISE AND",
			expr: "100 & 50",
			want: 32.0,
		},
		{
			name: "Single BITWISE OR",
			expr: "100 | 50",
			want: 118.0,
		},
		{
			name: "Single BITWISE XOR",
			expr: "100 ^ 50",
			want: 86.0,
		},
		{
			name: "Single shift left",
			expr: "2 << 1",
			want: 4.0,
		},
		{
			name: "Single shift right",
			expr: "2 >> 1",
			want: 1.0,
		},
		{

			name: "Single MULTIPLY",
			expr: "5 * 20",
			want: 100.0,
		},
		{

			name: "Single DIVIDE",
			expr: "100 / 20",
			want: 5.0,
		},
		{

			name: "Single even MODULUS",
			expr: "100 % 2",
			want: 0.0,
		},
		{
			name: "Single odd MODULUS",
			expr: "101 % 2",
			want: 1.0,
		},
		{

			name: "Compound PLUS",
			expr: "20 + 30 + 50",
			want: 100.0,
		},
		{

			name: "Compound BITWISE AND",
			expr: "20 & 30 & 50",
			want: 16.0,
		},
		{
			name: "Mutiple operators",
			expr: "20 * 5 - 49",
			want: 51.0,
		},
		{
			name: "Parenthesis usage",
			expr: "100 - (5 * 10)",
			want: 50.0,
		},
		{

			name: "Nested parentheses",
			expr: "50 + (5 * (15 - 5))",
			want: 100.0,
		},
		{

			name: "Nested parentheses with bitwise",
			expr: "100 ^ (23 * (2 | 5))",
			want: 197.0,
		},
		{
			name: "Logical OR operation of two clauses",
			expr: "(1 == 1) || (true == true)",
			want: true,
		},
		{
			name: "Logical AND operation of two clauses",
			expr: "(1 == 1) && (true == true)",
			want: true,
		},
		{

			name: "Implicit boolean",
			expr: "2 > 1",
			want: true,
		},
		{
			name: "Equal test minus numbers and no spaces",
			expr: "-1==-1",
			want: true,
		},
		{

			name: "Compound boolean",
			expr: "5 < 10 && 1 < 5",
			want: true,
		},
		{
			name: "Evaluated true && false operation (for issue #8)",
			expr: "1 > 10 && 11 > 10",
			want: false,
		},
		{

			name: "Evaluated true && false operation (for issue #8)",
			expr: "true == true && false == true",
			want: false,
		},
		{

			name: "Parenthesis boolean",
			expr: "10 < 50 && (1 != 2 && 1 > 0)",
			want: true,
		},
		{
			name: "Comparison of string constants",
			expr: `"foo" == "foo"`,
			want: true,
		},
		{
			name: "NEQ comparison of string constants",
			expr: `"foo" != "bar"`,
			want: true,
		},
		{

			name: "Multiplicative/additive order",
			expr: "5 + 10 * 2",
			want: 25.0,
		},
		{
			name: "Multiple constant multiplications",
			expr: "10 * 10 * 10",
			want: 1000.0,
		},
		{

			name: "Multiple adds/multiplications",
			expr: "10 * 10 * 10 + 1 * 10 * 10",
			want: 1100.0,
		},
		{

			name: "Modulus operatorPrecedence",
			expr: "1 + 101 % 2 * 5",
			want: 6.0,
		},
		{

			name: "Bit shift operatorPrecedence",
			expr: "50 << 1 & 90",
			want: 64.0,
		},
		{

			name: "Bit shift operatorPrecedence",
			expr: "90 & 50 << 1",
			want: 64.0,
		},
		{

			name: "Bit shift operatorPrecedence amongst non-bitwise",
			expr: "90 + 50 << 1 * 5",
			want: 4480.0,
		},
		{
			name: "Order of non-commutative same-operatorPrecedence operators (additive)",
			expr: "1 - 2 - 4 - 8",
			want: -13.0,
		},
		{
			name: "Order of non-commutative same-operatorPrecedence operators (multiplicative)",
			expr: "1 * 4 / 2 * 8",
			want: 16.0,
		},
		{
			name: "Sign prefix comparison",
			expr: "-1 < 0",
			want: true,
		},
		{

			name: "Lexicographic LT",
			expr: `"ab" < "abc"`,
			want: true,
		},
		{
			name: "Lexicographic LTE",
			expr: `"ab" <= "abc"`,
			want: true,
		},
		{

			name: "Lexicographic GT",
			expr: `"aba" > "abc"`,
			want: false,
		},
		{

			name: "Lexicographic GTE",
			expr: `"aba" >= "abc"`,
			want: false,
		},
		{

			name: "Boolean sign prefix comparison",
			expr: "!true == false",
			want: true,
		},
		{
			name: "Inversion of clause",
			expr: "!(10 < 0)",
			want: true,
		},
		{

			name: "Negation after modifier",
			expr: "10 * -10",
			want: -100.0,
		},
		{
			name: "String to string concat",
			expr: `"foo" + "bar" == "foobar"`,
			want: true,
		},
		{
			name: "String to float64 concat",
			expr: `"foo" + 123 == "foo123"`,
			want: true,
		},
		{

			name: "Float64 to string concat",
			expr: `123 + "bar" == "123bar"`,
			want: true,
		},
		{
			name: "Logical operator reordering (#30)",
			expr: "(true && true) || (true && false)",
			want: true,
		},
		{

			name: "Logical operator reordering without parens (#30)",
			expr: "true && true || true && false",
			want: true,
		},
		{

			name: "Logical operator reordering with multiple OR (#30)",
			expr: "false || true && true || false",
			want: true,
		},
		{
			name: "Left-side multiple consecutive (should be reordered) operators",
			expr: "(10 * 10 * 10) > 10",
			want: true,
		},
		{
			name: "Three-part non-paren logical op reordering (#44)",
			expr: "false && true || true",
			want: true,
		},
		{
			name: "Three-part non-paren logical op reordering (#44), second one",
			expr: "true || false && true",
			want: true,
		},
		{
			name: "Logical operator reordering without parens (#45)",
			expr: "true && true || false && false",
			want: true,
		},
	}

	for _, noParamsCase := range noParamsCases {
		t.Run(noParamsCase.name, func(t *testing.T) {
			got, err := Do(context.Background(), noParamsCase.expr, nil)
			if err != nil {
				t.Fatalf("Do() Name=%s expre=%s expected err nil but got %+v", noParamsCase.name, noParamsCase.expr, err)
				return
			} else if !reflect.DeepEqual(got, noParamsCase.want) {
				t.Fatalf("Do() Name=%s expre=%s expected %v but got %v", noParamsCase.name, noParamsCase.expr, noParamsCase.want, got)
			}

			parser, err := New(context.Background(), noParamsCase.expr)
			if err != nil {
				t.Fatalf("New() Name=%s expre=%s expected err nil but got %+v", noParamsCase.name, noParamsCase.expr, err)
				return
			}

			got, err = parser.Eval(nil)
			if err != nil {
				t.Fatalf("Eval() Name=%s expre=%s expected err nil but got %+v", noParamsCase.name, noParamsCase.expr, err)
				return
			} else if !reflect.DeepEqual(got, noParamsCase.want) {
				t.Fatalf("Eval() Name=%s expre=%s expected %v but got %v", noParamsCase.name, noParamsCase.expr, noParamsCase.want, got)
			}
		})
	}

}

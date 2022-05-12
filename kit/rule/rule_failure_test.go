package rule

import (
	"context"
	"fmt"
	"testing"
)

func Test_Failure(t *testing.T) {
	var failureCases = []struct {
		name string
		expr string
	}{
		{
			name: "Invalid equality comparator",
			expr: "1 = 1",
		},
		{
			name: "Invalid equality comparator",
			expr: "1 === 1",
		},
		{
			name: "Too many characters for logical operator",
			expr: "true &&& false",
		},
		{

			name: "Too many characters for logical operator",
			expr: "true ||| false",
		},
		{

			name: "Premature end to expression, via modifier",
			expr: "10 > 5 +",
		},
		{
			name: "Premature end to expression, via comparator",
			expr: "10 + 5 >",
		},
		{
			name: "Premature end to expression, via logical operator",
			expr: "10 > 5 &&",
		},
		{

			name: "Premature end to expression, via ternary operator",
			expr: "true ?",
		},
		{
			name: "Hanging REQ",
			expr: "`wat` =~",
		},
		{

			name: "Invalid operator change to REQ",
			expr: " / =~",
		},
		{
			name: "Invalid starting token, comparator",
			expr: "> 10",
		},
		{
			name: "Invalid starting token, modifier",
			expr: "+ 5",
		},
		{
			name: "Invalid starting token, logical operator",
			expr: "&& 5 < 10",
		},
		{
			name: "Invalid NUMERIC transition",
			expr: "10 10",
		},
		{
			name: "Invalid STRING transition",
			expr: "`foo` `foo`",
		},
		{
			name: "Invalid operator transition",
			expr: "10 > < 10",
		},
		{

			name: "Starting with unbalanced parens",
			expr: " ) ( arg2",
		},
		{

			name: "Unclosed bracket",
			expr: "[foo bar",
		},
		{

			name: "Unclosed quote",
			expr: "foo == `responseTime",
		},
		{

			name: "Constant regex pattern fail to compile",
			expr: "foo =~ `[abc`",
		},
		{

			name: "Constant unmatch regex pattern fail to compile",
			expr: "foo !~ `[abc`",
		},
		{

			name: "Unbalanced parentheses",
			expr: "10 > (1 + 50",
		},
		{

			name: "Multiple radix",
			expr: "127.0.0.1",
		},
		{

			name: "Hanging accessor",
			expr: "foo.Bar.",
		},
		{
			name: "Incomplete Hex",
			expr: "0x",
		},
		{
			name: "Invalid Hex literal",
			expr: "0x > 0",
		},
		{
			name: "Hex float (Unsupported)",
			expr: "0x1.1",
		},
		{
			name: "Hex invalid letter",
			expr: "0x12g1",
		},
		{
			name: "Error after camouflage",
			expr: "0 + ,",
		},
		{
			name: "Double func keyword",
			expr: "func func foo()",
		},
		{
			name: "Unknown func",
			expr: "func foo()",
		},
		{
			name: "Double func call paren",
			expr: "func in(1,1)()",
		},
		{
			name: "Unsupported {",
			expr: "{}",
		},
		{
			name: "Unsupported [",
			expr: "[]",
		},
		{
			name: "Unsupported $",
			expr: "1$1",
		},
	}

	for _, failureCase := range failureCases {
		t.Run(failureCase.name, func(t *testing.T) {
			ctx := context.Background()

			// do check
			if err := Check(ctx, failureCase.expr); err == nil {
				t.Fatalf("Check() Name=%s expre=%s expected error but got %v", failureCase.name, failureCase.expr, err)
				return
			}

			got, err := Do(ctx, failureCase.expr, nil)
			if err == nil {
				t.Fatalf("Do() Name=%s expre=%s expected error but got %v", failureCase.name, failureCase.expr, got)
				return
			} else {
				fmt.Printf("Do() Name=%s expre=%s err=%+v \n", failureCase.name, failureCase.expr, err)
			}

			err = nil

			_, err = New(ctx, failureCase.expr)
			if err == nil {
				t.Fatalf("New() Name=%s expre=%s expected error but got %v", failureCase.name, failureCase.expr, got)
				return
			} else {
				fmt.Printf("New() Name=%s expre=%s err=%+v \n", failureCase.name, failureCase.expr, err)
			}

		})
	}
}

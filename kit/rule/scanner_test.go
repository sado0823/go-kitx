package rule

import (
	"context"
	"fmt"
	"go/ast"
	buildinParser "go/parser"
	goscanner "go/scanner"
	"go/token"
	"strings"
	"testing"
	"text/scanner"
)


func Test_AST(t *testing.T) {
	expr := `(a-b >= 6) && true`

	// 通过解析src来创建AST。
	f, err := buildinParser.ParseExpr(expr)
	if err != nil {
		panic(err)
	}

	// 检查AST并打印所有标识符和文字。
	a := 0
	ast.Inspect(f, func(n ast.Node) bool {
		a++
		fmt.Printf("type=%T \n", n)
		var s string
		var s2 string
		switch x := n.(type) {
		case *ast.BasicLit:
			s = x.Value
			s2 = x.Kind.String()
		case *ast.Ident:
			s = x.Name
			s2 = x.String()
		case *ast.UnaryExpr:
			s = x.Op.String()
		case *ast.BinaryExpr:
			s = x.Op.String()
		}
		if s != "" {
			fmt.Printf("v=%s token=%s \t\n", s, s2)
		}
		return true
	})

	fmt.Println("a=", a)

}

func Test_parse(t *testing.T) {
	expr := `(foo - 90 > 0 ) && ( foo > 1 || foo <1 ) && foo > 1`
	param := map[string]interface{}{
		"foo": 100,
	}
	p := newParser(context.Background(), expr)
	tokens, err := p.read()
	if err != nil {
		panic(err)
	}
	for _, token := range tokens {
		fmt.Printf("pos_o=%d  pos=%s\t token=%s\t lit=%q\n", token.Pos(), p.fSet.Position(token.Pos()), token.Peer(), token.Lit())
	}
	stage, err := buildStageFromTokens(tokens)
	if err != nil {
		panic(err)
	}
	res, err := doStage(stage, param)
	if err != nil {
		panic(err)
	}
	fmt.Printf("res=%v\t type=%T err=%+v \n", res, res, err)
}

func Test_C(t *testing.T) {
	expr := `||`
	p := newParser(context.Background(), expr)
	tokens, err := p.read()
	if err != nil {
		panic(err)
	}
	for _, token := range tokens {
		fmt.Printf("pos_o=%d  pos=%s\t token=%s\t lit=%q\n", token.Pos(), p.fSet.Position(token.Pos()), token.Peer(), token.Lit())
	}
}

func Test_B(t *testing.T) {
	// src is the input that we want to tokenize.
	src := []byte(`;a >= 6 && a != "abc" && a.ABC() != 0;; // Euler.`)
	//src := []byte(`func A() int64 { return 1 } // Euler.`)

	// Initialize the scanner.
	var s goscanner.Scanner
	fset := token.NewFileSet()                      // positions are relative to fset
	file := fset.AddFile("", fset.Base(), len(src)) // register input "file"
	s.Init(file, src, nil /* no error handler */, goscanner.ScanComments)

	// Repeated calls to Scan yield the token sequence found in the input.
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.SEMICOLON {
			fmt.Println("有分号了")
		}
		fmt.Printf("pos:%s\t token:%s\t lit:%q\n", fset.Position(pos), tok, lit)
	}
}

func Test_A(t *testing.T) {
	exp := "a >= 6"

	var s scanner.Scanner
	s.Init(strings.NewReader(exp))

	s.Filename = exp + "\t"
	s.Error = func(s *scanner.Scanner, msg string) {
		fmt.Println("Scanner err msg: ", msg)
	}

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		sT := s.TokenText()
		fmt.Println("token text: ", sT)
	}
}

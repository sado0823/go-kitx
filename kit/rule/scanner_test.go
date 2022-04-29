package rule

import (
	"context"
	"fmt"
	goscanner "go/scanner"
	"go/token"
	"strings"
	"testing"
	"text/scanner"
)

func Test_Build_node(t *testing.T) {
	expr := `a-b >= 6 && b != 100 && c < 1; // Euler.; a>=b`
	p := newParser(context.Background(), expr)
	tokens := p.read()
	for _, token := range tokens {
		fmt.Printf("pos_o=%d  pos=%s\t token=%s\t token_str=%s\t lit=%q\n", token.pos, p.fSet.Position(token.pos), token.tok, token.tokStr, token.lit)
	}

	params := map[string]interface{}{
		"a": 10,
		"b": 2,
		"c": 3,
	}
	last := &last{lastNodes: make([]*node, 0)}
	node := buildNode(tokens, last)
	eval, err := node.Eval(last, params)
	fmt.Printf("eval=%v \t type=%T \t err=%+v \n", eval, eval, err)

	//fmt.Println(node, last)
}

func Test_C(t *testing.T) {
	expr := `;a >= 6 && a != "abc" && a.ABC() != 0; // Euler.; a>=b`
	p := newParser(context.Background(), expr)
	tokens := p.read()
	for _, token := range tokens {
		fmt.Printf("pos_o=%d  pos=%s\t token=%s\t lit=%q\n", token.pos, p.fSet.Position(token.pos), token.tok, token.lit)
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

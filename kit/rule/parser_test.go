package rule

import (
	"context"
	goscanner "go/scanner"
	"go/token"
	"strings"
	"testing"
	"text/scanner"
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
	expr := `foo.bar.Hobby.0.Name`
	param := map[string]interface{}{
		"foo": map[string]interface{}{
			"bar": T{Name: "tom", Hobby: []T{{Name: "jay"}}},
		},
		"in": 12.2,
	}
	parser, err := New(context.Background(), expr, WithCustomFn("test", func(arguments ...interface{}) (interface{}, error) {
		logger.Println("i am test func")
		logger.Println(arguments...)
		return true, nil
	}))
	if err != nil {
		panic(err)
	}
	res, err := parser.Eval(param)
	if err != nil {
		panic(err)
	}
	logger.Printf("res=%v\t type=%T\t err=%+v \n", res, res, err)
}

func Test_Do(t *testing.T) {
	//expr := `(foo - 90 > 0 ) && ( foo > 1 || foo <1 ) && foo > 1`
	expr := `1+101%2*5`
	param := map[string]interface{}{
		"foo": 5,
		"bar": 6,
	}

	res, err := Do(context.Background(), expr, param)
	if err != nil {
		panic(err)
	}
	logger.Printf("res=%v\t type=%T\t err=%+v \n", res, res, err)
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
			logger.Println("got SEMICOLON")
		}
		logger.Printf("pos:%s\t token:%s\t lit:%q\n", fset.Position(pos), tok, lit)
	}
}

func Test_A(t *testing.T) {
	exp := "a >= 6"

	var s scanner.Scanner
	s.Init(strings.NewReader(exp))

	s.Filename = exp + "\t"
	s.Error = func(s *scanner.Scanner, msg string) {
		logger.Println("Scanner err msg: ", msg)
	}

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		sT := s.TokenText()
		logger.Println("token text: ", sT)
	}
}

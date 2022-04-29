package rule

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/scanner"
	"go/token"
	"math"
	"strconv"
)

type parser struct {
	sc   scanner.Scanner
	fSet *token.FileSet
	file *token.File
}

type Token struct {
	pos    token.Pos
	tok    token.Token
	tokStr string
	lit    string
}

func newParser(ctx context.Context, expr string) *parser {
	src := []byte(expr)

	var s scanner.Scanner
	fSet := token.NewFileSet()
	file := fSet.AddFile("", fSet.Base(), len(src))
	s.Init(file, src, nil /* no error handler */, scanner.ScanComments)

	return &parser{sc: s, fSet: fSet, file: file}
}

func (p *parser) read() []*Token {
	tokens := make([]*Token, 0)
	for {
		pos, tok, lit := p.sc.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.SEMICOLON || tok == token.COMMENT {
			continue
		}
		tokens = append(tokens, &Token{
			pos:    pos,
			tok:    tok,
			tokStr: tok.String(),
			lit:    lit,
		})
	}
	return tokens
}

type stage struct {
	left  token.Token
	infix token.Token
	right token.Token
}

func (s *stage) full() bool {
	//IDENT  // main
	//INT    // 12345
	//FLOAT  // 123.45
	//STRING // "abc"
	toCheck := []token.Token{s.left, s.infix, s.right}
	for _, t := range toCheck {
		if !s.allow(t) {
			return false
		}
	}

	return true
}

func (s *stage) allow(tok token.Token) bool {
	allow := []token.Token{token.IDENT, token.INT, token.FLOAT, token.STRING}
	for _, t := range allow {
		if t == tok {
			return true
		}
	}

	return false
}

type node struct {
	symbol      token.Token
	leftTokens  []*Token
	rightTokens []*Token
	left        *node
	right       *node
	isLast      bool
}

type last struct {
	lastNodes []*node
	infix     []token.Token
}

func (n *node) buildExpr(tokens []*Token) string {
	var v bytes.Buffer
	for _, t := range tokens {
		if t.lit != "" {
			v.WriteString(t.lit + " ")
		}
		v.WriteString(t.tokStr + " ")
	}

	return v.String()
}

func (n *node) Eval(last *last, params map[string]interface{}) (interface{}, error) {
	var results []interface{}
	for _, lastNode := range last.lastNodes {
		compute, err := lastNode.compute(params)
		fmt.Printf("compute=%v \t err=%+v \n", compute, err)
		if err != nil {
			return nil, err
		}
		results = append(results, compute)
	}
	bools, err := parseBool(results...)
	if err != nil {
		return nil, err
	}
	var (
		left    = true
		infixes = []token.Token{token.LAND}
	)
	infixes = append(infixes, last.infix...)
	for _, result := range bools {

		infix := infixes[0]

		boolFn, ok := boolFuncMap[infix.String()]
		if !ok {
			return nil, fmt.Errorf("invalid bool func type=%s", infix.String())
		}

		left = boolFn(left, result)
		if err != nil {
			return nil, err
		}

		infixes = infixes[1:]
	}

	return left, nil
}

func (n *node) compute(params map[string]interface{}) (interface{}, error) {
	if !n.isLast {
		return nil, errors.New("not the last node")
	}

	if len(n.leftTokens) == 0 {
		return nil, errors.New("no left tokens")
	}

	//indexLength := len(n.leftTokens) - 1
	//for i, leftToken := range n.leftTokens {
	//	if leftToken.tok == token.SUB && i+1 > indexLength {
	//		return nil, fmt.Errorf("invalid expr:%s", n.buildExpr(n.leftTokens))
	//	}
	//}
	var (
		values []interface{}
		fn1    mathFunc
		fn2    compareFunc
	)
	for _, leftToken := range n.leftTokens {
		if leftToken.tok == token.IDENT {
			v, ok := params[leftToken.lit]
			if !ok {
				return nil, fmt.Errorf("invalid param:%s", leftToken.lit)
			}
			values = append(values, v)
		} else if mathFn, ok := mathFuncMap[leftToken.tokStr]; ok {
			fn1 = mathFn
		} else if compareFn, ok := compareFuncMap[leftToken.tokStr]; ok {
			fn2 = compareFn
		} else if leftToken.tok == token.FLOAT || leftToken.tok == token.INT || leftToken.tok == token.STRING {
			parseFloat, err := strconv.ParseFloat(leftToken.lit, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid param=%+v", leftToken)
			}
			values = append(values, parseFloat)
		}
	}

	fmt.Println(values)

	float64s, err := parseFloat64(values...)
	if err != nil {
		return nil, err
	}
	if fn1 != nil {
		return fn1(float64s[0], float64s[1]), nil
	}

	if fn2 != nil {
		return fn2(float64s[0], float64s[1]), nil
	}

	return nil, fmt.Errorf("unsupport fn=%s", n.buildExpr(n.leftTokens))
}

func buildNode(tokens []*Token, last *last) *node {
	if len(tokens) == 0 {
		return nil
	}
	for i, t := range tokens {
		if t.tok == token.LAND || t.tok == token.LOR {
			last.infix = append(last.infix, t.tok)
			node := &node{
				symbol:      t.tok,
				leftTokens:  tokens[0:i],
				rightTokens: tokens[i+1:],
				left:        nil,
				right:       nil,
			}

			nodeL := buildNode(node.leftTokens, last)

			nodeR := buildNode(node.rightTokens, last)

			node.left = nodeL
			node.right = nodeR

			return node
		}
	}

	lastNode := &node{
		leftTokens:  tokens,
		rightTokens: nil,
		left:        nil,
		right:       nil,
		isLast:      true,
	}
	last.lastNodes = append(last.lastNodes, lastNode)

	return lastNode
}

type mathFunc func(a, b float64) float64

var mathFuncMap = map[string]mathFunc{
	"+":  func(a, b float64) float64 { return a + b },
	"-":  func(a, b float64) float64 { return a - b },
	"*":  func(a, b float64) float64 { return a * b },
	"/":  func(a, b float64) float64 { return a / b },
	"%":  func(a, b float64) float64 { return math.Mod(a, b) },
	"**": func(a, b float64) float64 { return math.Pow(a, b) },
}

type boolFunc func(a, b bool) bool

var boolFuncMap = map[string]boolFunc{
	"&&": func(a, b bool) bool {
		return a && b
	},
	"||": func(a, b bool) bool {
		return a || b
	},
}

type compareFunc func(a, b float64) bool

var compareFuncMap = map[string]compareFunc{
	">":  func(a, b float64) bool { return a > b },
	">=": func(a, b float64) bool { return a >= b },
	"<":  func(a, b float64) bool { return a < b },
	"<=": func(a, b float64) bool { return a <= b },

	"==": func(a, b float64) bool { return a == b },
	"!=": func(a, b float64) bool { return a != b },
}

func parseFloat64(params ...interface{}) ([]float64, error) {
	var values []float64
	for _, param := range params {
		switch assert := param.(type) {
		case int, int8, int16, int32, int64, float32, float64, string:
			parseFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", assert), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid param=%v type=%T", param, param)
			}
			values = append(values, parseFloat)
		default:
			return nil, fmt.Errorf("invalid param=%v type=%T", param, param)
		}
	}
	return values, nil
}

func parseBool(params ...interface{}) ([]bool, error) {
	var values []bool
	for _, param := range params {

		switch assert := param.(type) {
		case int, int8, int16, int32, int64, float32, float64:
			parseFloat, err := strconv.ParseFloat(fmt.Sprintf("%v", assert), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid param=%v type=%T", param, param)
			}
			v := false
			if parseFloat > 0 {
				v = true
			}
			values = append(values, v)
		case string:
			parse, err := strconv.ParseBool(fmt.Sprintf("%v", assert))
			if err != nil {
				return nil, fmt.Errorf("invalid param=%v type=%T", param, param)
			}
			values = append(values, parse)
		case bool:
			values = append(values, assert)
		default:
			return nil, fmt.Errorf("invalid param=%v type=%T", param, param)
		}
	}
	return values, nil
}

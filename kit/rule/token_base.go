package rule

import (
	"fmt"
	"go/token"
)

type baseToken struct {
	pos    token.Pos
	tok    token.Token
	tokStr string
	lit    string
	value  interface{}
}

func (t *baseToken) canRunNext(allow []Symbol, next Token) error {
	for _, kind := range allow {
		if next.Symbol() == kind {
			return nil
		}
	}

	return fmt.Errorf("token(%s) can NOT exist after token(%s)", next.String(), t.String())
}

func (t *baseToken) Pos() token.Pos {
	return t.pos
}

func (t *baseToken) Peer() token.Token {
	return t.tok
}

func (t *baseToken) String() string {
	var s string
	if t.lit != "" {
		s = t.lit
	} else {
		s = t.tok.String()
	}
	return fmt.Sprintf("expr=%q \t pos=%v", s, t.pos)
}

func (t *baseToken) Lit() string {
	return t.lit
}

func (t *baseToken) Value() interface{} {
	return t.value
}

func (t *baseToken) LeftCheckFn() ParamCheckFn {
	return nil
}

func (t *baseToken) RightCheckFn() ParamCheckFn {
	return nil
}

func (t *baseToken) CanEOF() bool {
	return false
}

type comparableBase struct {
	baseToken
}

func (t *comparableBase) LeftCheckFn() ParamCheckFn {
	return func(left, right interface{}, param interface{}) error {
		_, ok := convertToFloat(left)
		if !ok {
			return fmt.Errorf("left should be a Number, but got %T, value=%v", left, left)
		}
		return nil
	}
}

func (t *comparableBase) RightCheckFn() ParamCheckFn {
	return func(left, right interface{}, param interface{}) error {
		_, ok := convertToFloat(right)
		if !ok {
			return fmt.Errorf("right should be a Number, but got %T, value=%v", right, right)
		}
		return nil
	}
}

func (t *comparableBase) CanNext(token Token) error {
	validNextKinds := []Symbol{
		NEGATE,
		NOT,
		NUMBER,
		BOOL,
		IDENT,
		FUNC,
		STRING,
		LPAREN,
		RPAREN,
	}

	return t.canRunNext(validNextKinds, token)
}

type boolBase struct {
	baseToken
}

func (t *boolBase) LeftCheckFn() ParamCheckFn {
	return func(left, right interface{}, param interface{}) error {
		switch left.(type) {
		case bool:
			return nil
		default:
			return fmt.Errorf("left should be bool, got:%T, TOKEN:%#v", left, t)
		}
	}
}

func (t *boolBase) RightCheckFn() ParamCheckFn {
	return func(left, right interface{}, param interface{}) error {
		switch right.(type) {
		case bool:
			return nil
		default:
			return fmt.Errorf("right should be bool, got:%T", right)
		}
	}
}

func (t *boolBase) CanNext(token Token) error {
	validNextKinds := []Symbol{
		IDENT,
		BOOL,
		FUNC,
		NEGATE,
		NUMBER,
		LPAREN,
		NOT,
	}

	return t.canRunNext(validNextKinds, token)
}

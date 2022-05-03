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

func (t *baseToken) Pos() token.Pos {
	return t.pos
}

func (t *baseToken) Peer() token.Token {
	return t.tok
}

func (t *baseToken) String() string {
	return fmt.Sprintf("token=%d\t token_string=%s\t lit=%s\t \n", t.tok, t.tok.String(), t.lit)
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

type comparableBase struct {
	baseToken
}

func (t *comparableBase) LeftCheckFn() ParamCheckFn {
	return func(left, right interface{}, param map[string]interface{}) error {
		_, l := left.(float64)
		if l {
			return nil
		}
		return fmt.Errorf("left should be float64, got:%T", left)
	}
}

func (t *comparableBase) RightCheckFn() ParamCheckFn {
	return func(left, right interface{}, param map[string]interface{}) error {
		_, r := right.(float64)
		if r {
			return nil
		}
		return fmt.Errorf("right should be float64, got:%T", right)
	}
}

type boolBase struct {
	baseToken
}

func (t *boolBase) LeftCheckFn() ParamCheckFn {
	return func(left, right interface{}, param map[string]interface{}) error {
		switch left.(type) {
		case bool:
			return nil
		default:
			return fmt.Errorf("left should be bool, got:%T", left)
		}
	}
}

func (t *boolBase) RightCheckFn() ParamCheckFn {
	return func(left, right interface{}, param map[string]interface{}) error {
		switch right.(type) {
		case bool:
			return nil
		default:
			return fmt.Errorf("right should be bool, got:%T", right)
		}
	}
}

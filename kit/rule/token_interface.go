package rule

import (
	"fmt"
	"go/token"
)

type (
	Symbol int64

	SymbolFn func(left, right interface{}, param map[string]interface{}) (interface{}, error)

	ParamCheckFn func(left, right interface{}, param map[string]interface{}) error

	Token interface {
		Pos() token.Pos
		Peer() token.Token
		String() string
		Lit() string
		Value() interface{}

		Symbol() Symbol
		SymbolFn() SymbolFn
		LeftCheckFn() ParamCheckFn
		RightCheckFn() ParamCheckFn
	}

	stage struct {
		symbol Token

		left  *stage
		right *stage
	}
)

func getIDENTFn(paramName string) SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		v, ok := param[paramName]
		if !ok {
			return nil, fmt.Errorf("IDENT param not found:%s", paramName)
		}
		if parse, ok := v.(int); ok {
			return float64(parse), nil
		}
		return v, nil
	}
}

func getLiteralFn(literal interface{}) SymbolFn {
	return func(left interface{}, right interface{}, parameters map[string]interface{}) (interface{}, error) {
		return literal, nil
	}
}

// param key
type tokenIdent struct {
	baseToken
}

func (t *tokenIdent) Symbol() Symbol {
	return Ident
}

func (t *tokenIdent) SymbolFn() SymbolFn {
	return getIDENTFn(t.lit)
}

// todo
// support func type
type tokenFunc struct {
	baseToken
}

func (t *tokenFunc) Symbol() Symbol {
	return Func
}

func (t *tokenFunc) SymbolFn() SymbolFn {
	return nil
}

// multi expr
type tokenSeparator struct {
	baseToken
}

func (t *tokenSeparator) Symbol() Symbol {
	return Comma
}

func (t *tokenSeparator) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		var ret []interface{}

		switch left.(type) {
		case []interface{}:
			ret = append(left.([]interface{}), right)
		default:
			ret = []interface{}{left, right}
		}

		return ret, nil
	}
}

// float64,int
type tokenNumber struct {
	baseToken
}

func (t *tokenNumber) Symbol() Symbol {
	return Number
}

func (t *tokenNumber) SymbolFn() SymbolFn {
	return getLiteralFn(t.value)
}

// string
type tokenString struct {
	baseToken
}

func (t *tokenString) Symbol() Symbol {
	return String
}

func (t *tokenString) SymbolFn() SymbolFn {
	return getLiteralFn(t.value)
}

// just as a tree node
type tokenNull struct {
	baseToken
}

func (t *tokenNull) Symbol() Symbol {
	return Null
}

func (t *tokenNull) SymbolFn() SymbolFn {
	return func(left interface{}, right interface{}, parameters map[string]interface{}) (interface{}, error) {
		return right, nil
	}
}



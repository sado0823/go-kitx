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
		Peer() token.Token // current token.Token
		String() string
		Lit() string
		Value() interface{}

		Symbol() Symbol
		SymbolFn() SymbolFn
		LeftCheckFn() ParamCheckFn
		RightCheckFn() ParamCheckFn

		CanNext(token Token) error
		CanEOF() bool
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

func (t *tokenIdent) CanNext(token Token) error {
	if token.Symbol() == LPAREN {
		return fmt.Errorf("invalid way to do func call, try `func %s()`", t.lit)
	}
	validNextKinds := []Symbol{
		EQL, // ==
		NEQ, // !=
		GTR, // >
		GEQ, // >=
		LSS, // <
		LEQ, // <=
		AND, // &
		OR,  // |
		XOR, // ^
		SHL, // <<
		SHR, // >>
		ADD, // +
		SUB, // -
		MUL, // *
		QUO, // /
		REM, // %
		NOT, // !
		Comma,
	}

	return t.canRunNext(validNextKinds, token)
}

func (t *tokenIdent) CanEOF() bool {
	return true
}

// todo
// support func type
type tokenFunc struct {
	baseToken
}

func (t *tokenFunc) Symbol() Symbol {
	return Func
}



func (t *tokenFunc) parseParams2Arr(arr []interface{}, v interface{}) []interface{} {
	switch assertV := v.(type) {
	case []interface{}:
		for _, subV := range assertV {
			arr = t.parseParams2Arr(arr, subV)
		}
	default:
		arr = append(arr, assertV)
	}
	return arr
}

func (t *tokenFunc) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {

		fn := _buildInCustomFn[t.value.(string)]

		if right == nil {
			return fn()
		}

		logger.Printf("func right %v, %T \n", right, right)

		params := make([]interface{}, 0)

		params = t.parseParams2Arr(params, right)
		return fn(params...)
	}
}

func (t *tokenFunc) CanNext(token Token) error {
	validNextKinds := []Symbol{
		Ident,
		LPAREN,
		RPAREN,
	}

	return t.canRunNext(validNextKinds, token)
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

func (t *tokenSeparator) CanNext(token Token) error {
	validNextKinds := []Symbol{
		NOT,    // !
		NEGATE, // -1,-2,-3...
		Number,
		Bool,
		String,
		Ident,
		Func,
		LPAREN, // (
		RPAREN, // )
	}

	return t.canRunNext(validNextKinds, token)
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

func (t *tokenNumber) CanNext(token Token) error {
	validNextKinds := []Symbol{
		EQL,    // ==
		NEQ,    // !=
		GTR,    // >
		GEQ,    // >=
		LSS,    // <
		LEQ,    // <=
		AND,    // &
		OR,     // |
		XOR,    // ^
		SHL,    // <<
		SHR,    // >>
		ADD,    // +
		SUB,    // -
		MUL,    // *
		QUO,    // /
		REM,    // %
		RPAREN, // )
		Comma,

		LAND, // &&
		LOR,  // ||
	}

	return t.canRunNext(validNextKinds, token)
}

func (t *tokenNumber) CanEOF() bool {
	return true
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

func (t *tokenString) CanNext(token Token) error {
	validNextKinds := []Symbol{
		Comma,  // ,
		EQL,    // ==
		NEQ,    // !=
		RPAREN, // )
	}

	return t.canRunNext(validNextKinds, token)
}

func (t *tokenString) CanEOF() bool {
	return true
}

type tokenBool struct {
	baseToken
}

func (t *tokenBool) Symbol() Symbol {
	return Bool
}

func (t *tokenBool) SymbolFn() SymbolFn {
	return getLiteralFn(t.value)
}

func (t *tokenBool) CanNext(token Token) error {
	validNextKinds := []Symbol{
		EQL,    // ==
		NEQ,    // !=
		Comma,  // ,
		LAND,   // &&
		LOR,    // ||
		RPAREN, // )
	}

	return t.canRunNext(validNextKinds, token)
}

func (t *tokenBool) CanEOF() bool {
	return true
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

func (t *tokenNull) CanNext(token Token) error {
	validNextKinds := []Symbol{
		NOT,    // !
		NEGATE, // -1,-2,-3...
		Number,
		Bool,
		Ident,
		Func,
		String,
		LPAREN,
	}

	return t.canRunNext(validNextKinds, token)
}

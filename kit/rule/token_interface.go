package rule

import (
	"context"
	"fmt"
	"go/token"
	"strings"

	"github.com/sado0823/go-kitx/pkg/collection/reflect"
)

type (
	Symbol int64

	SymbolFn func(left, right interface{}, param interface{}) (interface{}, error)

	ParamCheckFn func(left, right interface{}, param interface{}) error

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
)

func getLiteralFn(literal interface{}) SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		return literal, nil
	}
}

// param key
type tokenIdent struct {
	baseToken
}

func (t *tokenIdent) Symbol() Symbol {
	if index := strings.LastIndex(t.lit, "."); index == len(t.lit)-1 {
		return PERIOD
	}
	return IDENT
}

func (t *tokenIdent) SymbolFn() SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		var (
			selection interface{} = param
			keys                  = strings.Split(t.lit, ".")
			ok        bool
		)

		for _, key := range keys {
			selection, ok = reflect.PathSelect(context.Background(), key, selection)
			if !ok {
				return nil, fmt.Errorf("IDENT param(%s) NOT FOUND", t.lit)
			}
		}

		if parse, ok := convertToFloat(selection); ok {
			return parse, nil
		}

		return selection, nil
	}
}

func (t *tokenIdent) CanNext(token Token) error {
	if token.Symbol() == LPAREN {
		if strings.Contains(t.lit, ".") {
			return fmt.Errorf("unsupported accessor func call:%s()", t.lit)
		}
		return fmt.Errorf("invalid way to do func call, try `func %s()`", t.lit)
	}
	if t.Symbol() == PERIOD {
		return t.canRunNext([]Symbol{IDENT, NUMBER}, token)
	}
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
		NOT,    // !
		COMMA,  // ,
		LAND,   // &&
		LOR,    // ||
		PERIOD, // .
	}

	return t.canRunNext(validNextKinds, token)
}

func (t *tokenIdent) CanEOF() bool {
	return t.Symbol() != PERIOD
}

// todo
// support func type
type tokenFunc struct {
	baseToken
}

func (t *tokenFunc) Symbol() Symbol {
	return FUNC
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
	return func(left, right interface{}, param interface{}) (interface{}, error) {

		fn, ok := t.value.(CustomFn)
		if !ok || fn == nil {
			return nil, fmt.Errorf("invalid func, name=%s, type=%T, value=%v", t.lit, t.value, t.value)
		}

		wrap2Float := func(v interface{}, err error) (interface{}, error) {
			if err != nil {
				return v, err
			}
			if parse, ok := convertToFloat(v); ok {
				return parse, nil
			}
			return v, nil
		}

		if right == nil {
			return wrap2Float(fn(param))
		}

		logger.Printf("func right %v, %T \n", right, right)

		params := make([]interface{}, 0)

		params = t.parseParams2Arr(params, right)
		return wrap2Float(fn(param, params...))
	}
}

func (t *tokenFunc) CanNext(token Token) error {
	validNextKinds := []Symbol{
		IDENT,
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
	return COMMA
}

func (t *tokenSeparator) SymbolFn() SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		var ret []interface{}

		switch left := left.(type) {
		case []interface{}:
			ret = append(left, right)
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
		NUMBER,
		BOOL,
		STRING,
		IDENT,
		FUNC,
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
	return NUMBER
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
		COMMA,

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
	return STRING
}

func (t *tokenString) SymbolFn() SymbolFn {
	return getLiteralFn(t.value)
}

func (t *tokenString) CanNext(token Token) error {
	validNextKinds := []Symbol{
		COMMA,  // ,
		EQL,    // ==
		NEQ,    // !=
		RPAREN, // )
		ADD,    // +
		LAND,   // &&
		LOR,    // ||
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
	return BOOL
}

func (t *tokenBool) SymbolFn() SymbolFn {
	return getLiteralFn(t.value)
}

func (t *tokenBool) CanNext(token Token) error {
	validNextKinds := []Symbol{
		EQL,    // ==
		NEQ,    // !=
		COMMA,  // ,
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
	return NULL
}

func (t *tokenNull) SymbolFn() SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		return right, nil
	}
}

func (t *tokenNull) CanNext(token Token) error {
	validNextKinds := []Symbol{
		NOT,    // !
		NEGATE, // -1,-2,-3...
		NUMBER,
		BOOL,
		IDENT,
		FUNC,
		STRING,
		LPAREN,
	}

	return t.canRunNext(validNextKinds, token)
}

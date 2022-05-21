package rule

import (
	"fmt"
	"reflect"
)

// ==
type tokenEQL struct {
	comparableBase
}

func (t *tokenEQL) Symbol() Symbol {
	return EQL
}

func (t *tokenEQL) LeftCheckFn() ParamCheckFn {
	return func(left, right interface{}, param interface{}) error {
		leftT, rightT, ok := typeEqual(left, right)
		if !ok {
			return fmt.Errorf("tokenEQL left type=%s should be equal to right type=%s", leftT.String(), rightT)
		}
		return nil
	}
}

func (t *tokenEQL) RightCheckFn() ParamCheckFn {
	return func(left, right interface{}, param interface{}) error {
		return nil
	}
}

func (t *tokenEQL) SymbolFn() SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		return reflect.DeepEqual(left, right), nil
	}
}

// !=
type tokenNEQ struct {
	comparableBase
}

func (t *tokenNEQ) Symbol() Symbol {
	return NEQ
}

func (t *tokenNEQ) LeftCheckFn() ParamCheckFn {
	return func(left, right interface{}, param interface{}) error {
		leftT, rightT, ok := typeEqual(left, right)
		if !ok {
			return fmt.Errorf("tokenEQL left type=%s should be equal to right type=%s", leftT.String(), rightT)
		}
		return nil
	}
}

func (t *tokenNEQ) RightCheckFn() ParamCheckFn {
	return func(left, right interface{}, param interface{}) error {
		return nil
	}
}

func (t *tokenNEQ) SymbolFn() SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		return !reflect.DeepEqual(left, right), nil
	}
}

// >
type tokenGTR struct {
	comparableBase
}

func (t *tokenGTR) Symbol() Symbol {
	return GTR
}

func (t *tokenGTR) SymbolFn() SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		return left.(float64) > right.(float64), nil
	}
}

// >=
type tokenGEQ struct {
	comparableBase
}

func (t *tokenGEQ) Symbol() Symbol {
	return GEQ
}

func (t *tokenGEQ) SymbolFn() SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		return left.(float64) >= right.(float64), nil
	}
}

// <
type tokenLSS struct {
	comparableBase
}

func (t *tokenLSS) Symbol() Symbol {
	return LSS
}

func (t *tokenLSS) SymbolFn() SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		return left.(float64) < right.(float64), nil
	}
}

// <=
type tokenLEQ struct {
	comparableBase
}

func (t *tokenLEQ) Symbol() Symbol {
	return LEQ
}

func (t *tokenLEQ) SymbolFn() SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		return left.(float64) <= right.(float64), nil
	}
}

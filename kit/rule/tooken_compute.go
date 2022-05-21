package rule

import (
	"fmt"
	"math"
)

// +
type tokenADD struct {
	comparableBase
}

func (t *tokenADD) Symbol() Symbol {
	return ADD
}

func (t *tokenADD) LeftCheckFn() ParamCheckFn {
	return func(left, right interface{}, param interface{}) error {
		_, ok := convertToFloat(left)
		if !isString(left) && !ok {
			return fmt.Errorf("add left should be a Number or String, but got %T, value=%v", left, left)
		}
		return nil
	}
}

func (t *tokenADD) RightCheckFn() ParamCheckFn {
	return func(left, right interface{}, param interface{}) error {
		_, ok := convertToFloat(right)
		if !isString(right) && !ok {
			return fmt.Errorf("add right should be a Number or String, but got %T, value=%v", right, right)
		}
		return nil
	}
}

func (t *tokenADD) SymbolFn() SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		if isString(left) || isString(right) {
			return fmt.Sprintf("%v%v", left, right), nil
		}

		return left.(float64) + right.(float64), nil
	}
}

// -
type tokenSUB struct {
	comparableBase
}

func (t *tokenSUB) Symbol() Symbol {
	return SUB
}

func (t *tokenSUB) SymbolFn() SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		return left.(float64) - right.(float64), nil
	}
}

// *
type tokenMUL struct {
	comparableBase
}

func (t *tokenMUL) Symbol() Symbol {
	return MUL
}

func (t *tokenMUL) SymbolFn() SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		return left.(float64) * right.(float64), nil
	}
}

// /
type tokenQUO struct {
	comparableBase
}

func (t *tokenQUO) Symbol() Symbol {
	return QUO
}

func (t *tokenQUO) SymbolFn() SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		return left.(float64) / right.(float64), nil
	}
}

// %
type tokenREM struct {
	comparableBase
}

func (t *tokenREM) Symbol() Symbol {
	return REM
}

func (t *tokenREM) SymbolFn() SymbolFn {
	return func(left, right interface{}, param interface{}) (interface{}, error) {
		return math.Mod(left.(float64), right.(float64)), nil
	}
}

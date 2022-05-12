package rule

import "fmt"



// +
type tokenADD struct {
	comparableBase
}

func (t *tokenADD) Symbol() Symbol {
	return ADD
}

func (t *tokenADD) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		l1, ok1 := left.(float64)
		r1, ok2 := right.(float64)
		if ok1 && ok2 {
			return l1 + r1, nil
		}

		l2, ok1 := left.(string)
		r2, ok2 := right.(string)
		if ok1 && ok2 {
			return l2 + r2, nil
		}

		return nil, fmt.Errorf("tokenADD unsupported type to do add, left=%v,right=%v", left, right)
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
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
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
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
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
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
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
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		return float64(uint64(left.(float64)) % uint64(right.(float64))), nil
	}
}

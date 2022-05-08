package rule

import "fmt"

// ==
type tokenEQL struct {
	comparableBase
}

func (t *tokenEQL) Symbol() Symbol {
	return EQL
}

func (t *tokenEQL) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		l1, ok1 := left.(string)
		r1, ok2 := right.(string)
		if ok1 && ok2 {
			return l1 == r1, nil
		}

		l2, ok1 := left.(float64)
		r2, ok2 := right.(float64)
		if ok1 && ok2 {
			return l2 == r2, nil
		}

		l3, ok1 := left.(bool)
		r3, ok2 := right.(bool)
		if ok1 && ok2 {
			return l3 == r3, nil
		}

		return nil, fmt.Errorf("invalid left=%v, or right=%v", left, right)
	}
}

// !=
type tokenNEQ struct {
	comparableBase
}

func (t *tokenNEQ) Symbol() Symbol {
	return NEQ
}

func (t *tokenNEQ) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		return left.(float64) != right.(float64), nil
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
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
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
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
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
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
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
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		return left.(float64) <= right.(float64), nil
	}
}

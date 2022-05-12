package rule

import "fmt"

// !
type tokenNOT struct {
	baseToken
}

func (t *tokenNOT) RightCheckFn() ParamCheckFn {
	return func(left, right interface{}, param map[string]interface{}) error {
		switch right.(type) {
		case bool:
			return nil
		default:
			return fmt.Errorf("right should be bool, got:%T", right)
		}
	}
}

func (t *tokenNOT) Symbol() Symbol {
	return NOT
}

func (t *tokenNOT) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		return !right.(bool), nil
	}
}

func (t *tokenNOT) CanNext(token Token) error {
	validNextKinds := []Symbol{
		Ident,
		Bool,
		Func,
		LPAREN,
		NOT,
	}

	return t.canRunNext(validNextKinds, token)
}

type tokenNEGATE struct {
	baseToken
}

func (t *tokenNEGATE) Symbol() Symbol {
	return NEGATE
}

func (t *tokenNEGATE) RightCheckFn() ParamCheckFn {
	return func(left, right interface{}, param map[string]interface{}) error {
		_, ok1 := right.(int)
		_, ok2 := right.(float64)
		if !ok1 && !ok2 {
			return fmt.Errorf("tokenNEGATE right should be int or float64, got:%T, value:%v", right, right)

		}
		return nil
	}
}

func (t *tokenNEGATE) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		return -right.(float64), nil
	}
}

func (t *tokenNEGATE) CanNext(token Token) error {
	validNextKinds := []Symbol{
		Number,
		Func,
		LPAREN,
	}

	return t.canRunNext(validNextKinds, token)
}

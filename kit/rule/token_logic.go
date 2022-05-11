package rule

// !
type tokenNOT struct {
	boolBase
}

func (t *tokenNOT) LeftCheckFn() ParamCheckFn {
	return func(left, right interface{}, param map[string]interface{}) error {
		return nil
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
		Number,
		LPAREN,
		NOT,
	}

	return t.canRunNext(validNextKinds, token)
}

// ||
type tokenLOR struct {
	boolBase
}

func (t *tokenLOR) Symbol() Symbol {
	return LOR
}

func (t *tokenLOR) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		return left.(bool) || right.(bool), nil
	}
}

// &&
type tokenLAND struct {
	boolBase
}

func (t *tokenLAND) Symbol() Symbol {
	return LAND
}

func (t *tokenLAND) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		return left.(bool) && right.(bool), nil
	}
}

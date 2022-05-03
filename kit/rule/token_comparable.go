package rule

// ==
type tokenEQL struct {
	comparableBase
}

func (t *tokenEQL) Symbol() Symbol {
	return EQL
}

func (t *tokenEQL) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		return left.(float64) == right.(float64), nil
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

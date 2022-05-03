package rule

type tokenNEGATE struct {
	comparableBase
}

func (t *tokenNEGATE) Symbol() Symbol {
	return NEGATE
}

func (t *tokenNEGATE) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		return -right.(float64), nil
	}
}

// +
type tokenADD struct {
	comparableBase
}

func (t *tokenADD) Symbol() Symbol {
	return ADD
}

func (t *tokenADD) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
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

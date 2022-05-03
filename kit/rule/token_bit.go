package rule

// &
type tokenAND struct {
	comparableBase
}

func (t *tokenAND) Symbol() Symbol {
	return AND
}

func (t *tokenAND) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		return float64(int64(left.(float64)) & int64(right.(float64))), nil
	}
}

// |
type tokenOR struct {
	comparableBase
}

func (t *tokenOR) Symbol() Symbol {
	return OR
}

func (t *tokenOR) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		return float64(int64(left.(float64)) | int64(right.(float64))), nil
	}
}

// ^
type tokenXOR struct {
	comparableBase
}

func (t *tokenXOR) Symbol() Symbol {
	return XOR
}

func (t *tokenXOR) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		return float64(int64(left.(float64)) ^ int64(right.(float64))), nil
	}
}

// <<
type tokenSHL struct {
	comparableBase
}

func (t *tokenSHL) Symbol() Symbol {
	return SHL
}

func (t *tokenSHL) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		return float64(uint64(left.(float64)) << uint64(right.(float64))), nil
	}
}

// >>
type tokenSHR struct {
	comparableBase
}

func (t *tokenSHR) Symbol() Symbol {
	return SHR
}

func (t *tokenSHR) SymbolFn() SymbolFn {
	return func(left, right interface{}, param map[string]interface{}) (interface{}, error) {
		return float64(uint64(left.(float64)) >> uint64(right.(float64))), nil
	}
}

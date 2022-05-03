package rule

// will hack handle in valueStage
type tokenHackLPAREN struct {
	baseToken
}

func (t *tokenHackLPAREN) Symbol() Symbol {
	return LPAREN
}

func (t *tokenHackLPAREN) SymbolFn() SymbolFn {
	return nil
}

// will hack handle in valueStage
type tokenHackRPAREN struct {
	baseToken
}

func (t *tokenHackRPAREN) Symbol() Symbol {
	return RPAREN
}

func (t *tokenHackRPAREN) SymbolFn() SymbolFn {
	return nil
}

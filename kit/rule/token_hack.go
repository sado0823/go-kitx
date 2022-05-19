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

func (t *tokenHackLPAREN) CanNext(token Token) error {
	validNextKinds := []Symbol{
		NOT,    // !
		NEGATE, // -1,-2,-3...
		NUMBER,
		BOOL,
		IDENT,
		FUNC,
		STRING,
		LPAREN,
		RPAREN,
	}

	return t.canRunNext(validNextKinds, token)
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

func (t *tokenHackRPAREN) CanNext(token Token) error {
	validNextKinds := []Symbol{
		EQL, // ==
		NEQ, // !=
		GTR, // >
		GEQ, // >=
		LSS, // <
		LEQ, // <=
		AND, // &
		OR,  // |
		XOR, // ^
		SHL, // <<
		SHR, // >>
		ADD, // +
		SUB, // -
		MUL, // *
		QUO, // /
		REM, // %
		NOT, // !

		LAND, // &&
		LOR,  // ||

		NUMBER,
		BOOL,
		IDENT,
		STRING,
		RPAREN,
		COMMA,
	}

	return t.canRunNext(validNextKinds, token)
}

func (t *tokenHackRPAREN) CanEOF() bool {
	return true
}

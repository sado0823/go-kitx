package rule

import (
	"go/token"
	"strconv"
)

const (
	Null Symbol = iota
	Ident
	Bool
	Func
	EQL // ==
	NEQ // !=
	GTR // >
	GEQ // >=
	LSS // <
	LEQ // <=
	AND // &
	OR  // |
	XOR // ^
	SHL // <<
	SHR // >>
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %
	NOT // !
	Number
	String
	FLOAT
	Comma  // ,
	LAND   // &&
	LOR    // ||
	NEGATE // -1,-2,-3...
	LPAREN // (
	RPAREN // )
)

var token2Symbol = map[token.Token]Symbol{
	token.IDENT:  Ident,
	token.INT:    Number,
	token.FLOAT:  Number,
	token.STRING: String,

	token.LAND: LAND, // &&
	token.LOR:  LOR,  // ||

	token.ADD: ADD, // +
	token.SUB: SUB, // -
	token.MUL: MUL, // *
	token.QUO: QUO, // /
	token.REM: REM, // %

	token.AND: AND, // &
	token.OR:  OR,  // |
	token.XOR: XOR, // ^
	token.SHL: SHL, // <<
	token.SHR: SHR, // >>

	token.EQL: EQL, // ==
	token.LSS: LSS, // <
	token.GTR: GTR, // >
	token.NOT: NOT, // !

	token.NEQ: NEQ, // !=
	token.LEQ: LEQ, // <=
	token.GEQ: GEQ, // >=

	token.LPAREN: LPAREN, // (
	token.RPAREN: RPAREN, // )

	token.COMMA: Comma, // ,
}

var symbol2Token = map[Symbol]func(pos token.Pos, tok token.Token, lit string) (Token, error){
	Func: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenFunc{
			baseToken: baseToken{
				pos:    pos,
				tok:    tok,
				tokStr: tok.String(),
				lit:    lit,
				value:  lit,
			},
		}, nil
	},
	Ident: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenIdent{
			baseToken: baseToken{
				pos:    pos,
				tok:    tok,
				tokStr: tok.String(),
				lit:    lit,
				value:  lit,
			},
		}, nil
	},
	Comma: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenSeparator{
			baseToken: baseToken{
				pos:    pos,
				tok:    tok,
				tokStr: tok.String(),
				lit:    lit,
				value:  lit,
			},
		}, nil
	},
	LOR: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenLOR{
			boolBase: boolBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	LAND: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenLAND{
			boolBase: boolBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	EQL: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenEQL{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	NEQ: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenNEQ{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	GTR: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenGTR{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	GEQ: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenGEQ{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	LSS: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenLSS{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	LEQ: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenLEQ{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	AND: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenAND{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	OR: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenOR{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	XOR: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenXOR{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	SHL: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenSHL{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	SHR: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenSHR{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	ADD: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenADD{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	SUB: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenSUB{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	MUL: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenMUL{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	QUO: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenQUO{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	REM: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenREM{
			comparableBase: comparableBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	NOT: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenNOT{
			boolBase: boolBase{
				baseToken: baseToken{
					pos:    pos,
					tok:    tok,
					tokStr: tok.String(),
					lit:    lit,
					value:  lit,
				},
			},
		}, nil
	},
	Number: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		ret := new(tokenNumber)
		var value interface{}
		if tok == token.INT {
			parseInt, err := strconv.ParseInt(lit, 10, 64)
			if err != nil {
				return nil, err
			}
			value = float64(parseInt)
		}
		if tok == token.FLOAT {
			parseFloat, err := strconv.ParseFloat(lit, 64)
			if err != nil {
				return nil, err
			}
			value = parseFloat
		}
		ret.baseToken = baseToken{
			pos:    pos,
			tok:    tok,
			tokStr: tok.String(),
			lit:    lit,
			value:  value,
		}
		return ret, nil
	},
	Bool: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenBool{boolBase{
			baseToken: baseToken{
				pos:    pos,
				tok:    tok,
				tokStr: tok.String(),
				lit:    lit,
				value:  lit,
			},
		}}, nil
	},
	String: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenString{
			baseToken: baseToken{
				pos:    pos,
				tok:    tok,
				tokStr: tok.String(),
				lit:    lit,
				value:  lit,
			},
		}, nil
	},
	LPAREN: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenHackLPAREN{
			baseToken: baseToken{
				pos:    pos,
				tok:    tok,
				tokStr: tok.String(),
				lit:    lit,
				value:  lit,
			},
		}, nil
	},
	RPAREN: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenHackRPAREN{
			baseToken: baseToken{
				pos:    pos,
				tok:    tok,
				tokStr: tok.String(),
				lit:    lit,
				value:  lit,
			},
		}, nil
	},
}

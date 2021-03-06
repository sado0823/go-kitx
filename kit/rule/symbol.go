package rule

import (
	"fmt"
	"go/token"
	"strconv"
	"strings"
)

const (
	UNKNOWN Symbol = iota
	NULL

	literalBeg
	IDENT
	NUMBER
	STRING
	BOOL
	literalEnd

	keywordBeg
	FUNC
	keywordEnd

	operatorBeg

	comparatorBeg
	EQL // ==
	NEQ // !=
	GTR // >
	GEQ // >=
	LSS // <
	LEQ // <=
	comparatorEnd

	bitwiseBeg
	AND // &
	OR  // |
	XOR // ^
	bitwiseEnd

	bitwiseShiftBeg
	SHL // <<
	SHR // >>
	bitwiseShiftEnd

	additiveBeg
	ADD // +
	SUB // -
	additiveEnd

	multiplicativeBeg
	MUL // *
	QUO // /
	REM // %
	multiplicativeEnd

	prefixBeg
	NOT    // !
	NEGATE // -1,-2,-3...
	prefixEnd

	logicBeg
	LAND // &&
	LOR  // ||
	logicEnd

	LPAREN // (
	RPAREN // )

	COMMA  // ,
	PERIOD // .
	operatorEnd
)

func symbolPrecedence(symbol Symbol) int {
	switch symbol {
	case NULL:
		return 0
	case IDENT, NUMBER, STRING, BOOL:
		return 1
	case FUNC:
		return 2
	case NOT, NEGATE:
		return 3
	case LOR:
		return 4
	case LAND:
		return 5
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return 6
	case ADD, SUB:
		return 7
	case MUL, QUO, REM:
		return 8
	case AND, OR, XOR:
		return 9
	case SHL, SHR:
		return 10
	case COMMA:
		return 11
	}
	return 1
}

func isLiteral(symbol Symbol) bool { return literalBeg < symbol && symbol < literalEnd }

func isOperator(symbol Symbol) bool { return operatorBeg < symbol && symbol < operatorEnd }

func isKeyword(symbol Symbol) bool { return keywordBeg < symbol && symbol < keywordEnd }

var token2Symbol = map[token.Token]Symbol{
	token.FUNC:   FUNC,
	token.IDENT:  IDENT,
	token.INT:    NUMBER,
	token.FLOAT:  NUMBER,
	token.STRING: STRING,

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

	token.COMMA: COMMA, // ,
}

var symbol2Token = map[Symbol]func(pos token.Pos, tok token.Token, lit string) (Token, error){
	FUNC: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
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
	IDENT: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
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
	COMMA: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
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
			baseToken: baseToken{
				pos:    pos,
				tok:    tok,
				tokStr: tok.String(),
				lit:    lit,
				value:  lit,
			},
		}, nil
	},
	NEGATE: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		return &tokenNEGATE{
			baseToken: baseToken{
				pos:    pos,
				tok:    tok,
				tokStr: tok.String(),
				lit:    lit,
				value:  lit,
			},
		}, nil
	},
	NUMBER: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
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
	BOOL: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		var v bool
		if strings.ToUpper(lit) == "TRUE" {
			v = true
		}
		return &tokenBool{
			baseToken: baseToken{
				pos:    pos,
				tok:    tok,
				tokStr: tok.String(),
				lit:    lit,
				value:  v,
			},
		}, nil
	},
	STRING: func(pos token.Pos, tok token.Token, lit string) (Token, error) {
		v, err := strconv.Unquote(lit)
		if err != nil {
			return nil, fmt.Errorf("invalid string input, err=%+v, expr=%s, pos=%v", err, tok.String(), lit)
		}
		return &tokenString{
			baseToken: baseToken{
				pos:    pos,
				tok:    tok,
				tokStr: tok.String(),
				lit:    lit,
				value:  v,
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

package rule

type stream struct {
	tokens []Token
	index  int
	len    int
}

func (s *stream) rewind() {
	s.index -= 1
}

func (s *stream) next() Token {
	t := s.tokens[s.index]
	s.index += 1
	return t
}

func (s stream) hasNext() bool {
	return s.index < s.len
}

type (
	// operator level
	leveler func(stream *stream) (*stage, error)

	planner struct {
		// valid operators
		validTokens []Token

		nextLeft  leveler
		nextRight leveler
	}
)

var planPrefix leveler
var planMultiplicative leveler
var planAdditive leveler
var planBitwise leveler
var planShift leveler
var planComparator leveler
var planLogicalAnd leveler
var planLogicalOr leveler
var planSeparator leveler

func init() {
	// todo complete prefix type rule
	// 优先级越高的操作符, 在树的越下层, 极限就是叶子节点
	planPrefix = planer2level(&planner{
		validTokens: []Token{&tokenNOT{}, &tokenNEGATE{}},
		nextRight:   funcStage,
	})
	planMultiplicative = planer2level(&planner{
		validTokens: []Token{&tokenMUL{}, &tokenQUO{}, &tokenREM{}},
		nextLeft:    funcStage,
	})
	planAdditive = planer2level(&planner{
		validTokens: []Token{&tokenADD{}, &tokenSUB{}},
		nextLeft:    planMultiplicative,
	})
	planShift = planer2level(&planner{
		validTokens: []Token{&tokenSHL{}, &tokenSHR{}},
		nextLeft:    planAdditive,
	})
	planBitwise = planer2level(&planner{
		validTokens: []Token{&tokenAND{}, &tokenOR{}, &tokenXOR{}},
		nextLeft:    planShift,
	})
	planComparator = planer2level(&planner{
		validTokens: []Token{&tokenEQL{}, &tokenNEQ{}, &tokenGTR{}, &tokenGEQ{}, &tokenLSS{}, &tokenLEQ{}},
		nextLeft:    planBitwise,
	})
	planLogicalAnd = planer2level(&planner{
		validTokens: []Token{&tokenLAND{}},
		nextLeft:    planComparator,
	})
	planLogicalOr = planer2level(&planner{
		validTokens: []Token{&tokenLOR{}},
		nextLeft:    planLogicalAnd,
	})
	planSeparator = planer2level(&planner{
		validTokens: []Token{&tokenSeparator{}},
		nextLeft:    planLogicalOr,
	})
}

func planer2level(planner *planner) leveler {

	var generated leveler
	var nextRight leveler

	generated = func(stream *stream) (*stage, error) {
		return level2stage(
			stream,
			planner.validTokens,
			nextRight,
			planner.nextLeft,
		)
	}

	if planner.nextRight != nil {
		nextRight = planner.nextRight
	} else {
		nextRight = generated
	}

	return generated
}

func level2stage(stream *stream, validTokens []Token, right, left leveler) (*stage, error) {
	var tokenNow Token
	var leftStage, rightStage *stage
	var err error
	var keyFound bool

	if left != nil {

		leftStage, err = left(stream)
		if err != nil {
			return nil, err
		}
	}

	for stream.hasNext() {

		tokenNow = stream.next()

		if len(validTokens) > 0 {

			keyFound = false
			for _, validToken := range validTokens {
				if validToken.Symbol() == tokenNow.Symbol() {
					keyFound = true
					break
				}
			}

			if !keyFound {
				break
			}
		}

		if right != nil {
			rightStage, err = right(stream)
			if err != nil {
				return nil, err
			}
		}

		return &stage{
			symbol: tokenNow,
			left:   leftStage,
			right:  rightStage,
		}, nil
	}

	stream.rewind()
	return leftStage, nil
}

func funcStage(stream *stream) (*stage, error) {

	tokenNow := stream.next()

	if tokenNow.Symbol() != Func {
		stream.rewind()
		return valueStage(stream)
	}

	rightStage, err := valueStage(stream)
	if err != nil {
		return nil, err
	}

	return &stage{
		symbol: tokenNow,
		right:  rightStage,
	}, nil

}

func valueStage(stream *stream) (*stage, error) {

	var tokenNow Token
	var ret *stage
	var err error

	if !stream.hasNext() {
		return nil, nil
	}

	tokenNow = stream.next()

	switch tokenNow.Symbol() {

	case LPAREN:

		ret, err = buildStage(stream)
		if err != nil {
			return nil, err
		}

		// advance past the CLAUSE_CLOSE token. We know that it's a CLAUSE_CLOSE,
		// because at parse-time we check for unbalanced parens.
		stream.next()

		ret = &stage{
			right:  ret,
			symbol: &tokenNull{},
		}

		return ret, nil

	case RPAREN:

		// when functions have empty params, this will be hit. In this case, we don't have any evaluation stage to do,
		// so we just return nil so that the stage planner continues on its way.
		stream.rewind()
		return nil, nil

	case NOT, NEGATE:
		stream.rewind()
		return planPrefix(stream)

	default:
		return &stage{
			symbol: tokenNow,
		}, nil
	}

}

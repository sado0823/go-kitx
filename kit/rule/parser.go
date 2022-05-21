package rule

import (
	"context"
	"fmt"
	"go/scanner"
	"go/token"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var (
	logger = log.New(os.Stdout, fmt.Sprintf("[DEBUG][pkg=rule][%s] ", time.Now().Format(time.StampMilli)), log.Lshortfile)
)

func init() {
	logger.SetFlags(0)
	logger.SetOutput(io.Discard)
}

type (
	Parser struct {
		ctx context.Context

		sc   scanner.Scanner
		fSet *token.FileSet
		file *token.File

		expr    string
		streamV *stream
		stageV  *stage

		option *option
	}

	option struct {
		customFns map[string]CustomFn
	}

	WithOption func(*option)
)

func WithCustomFn(name string, fn CustomFn) WithOption {
	return func(o *option) {
		o.customFns[name] = fn
	}
}

func doOptions(options ...WithOption) *option {
	option := &option{
		customFns: make(map[string]CustomFn),
	}
	for _, withOption := range options {
		withOption(option)
	}

	return option
}

func Check(ctx context.Context, expr string, options ...WithOption) (err error) {
	src := []byte(expr)

	var s scanner.Scanner
	fSet := token.NewFileSet()
	file := fSet.AddFile("", fSet.Base(), len(src))
	s.Init(file, src, nil /* no error handler */, scanner.ScanComments)

	p := &Parser{ctx: ctx, sc: s, fSet: fSet, file: file, expr: expr, option: doOptions(options...)}
	p.streamV, err = p.stream()

	return err
}

func Do(ctx context.Context, expr string, params interface{}, options ...WithOption) (interface{}, error) {
	parser, err := New(ctx, expr, options...)
	if err != nil {
		return nil, err
	}

	return parser.Eval(params)
}

func New(ctx context.Context, expr string, options ...WithOption) (parser *Parser, err error) {
	src := []byte(expr)

	var s scanner.Scanner
	fSet := token.NewFileSet()
	file := fSet.AddFile("", fSet.Base(), len(src))
	s.Init(file, src, nil /* no error handler */, scanner.ScanComments)

	p := &Parser{ctx: ctx, sc: s, fSet: fSet, file: file, expr: expr, option: doOptions(options...)}
	p.streamV, err = p.stream()
	if err != nil {
		return nil, err
	}

	p.stageV, err = buildStage(p.streamV)

	return p, err
}

func (p *Parser) Eval(params interface{}) (interface{}, error) {
	return doStage(p.stageV, params)
}

func (p *Parser) stream() (ret *stream, err error) {
	ret = new(stream)
	ret.tokens, err = p.read()
	ret.len = len(ret.tokens)
	return ret, err
}

func (p *Parser) read() ([]Token, error) {
	tokens := make([]Token, 0)
	var (
		beforeToken Token = new(tokenNull)
		lParenCount       = 0
		rParenCount       = 0
	)
	for {
		pos, tok, lit := p.sc.Scan()
		logger.Printf("pos_o=%d  pos=%s\t token=%#v\t token_str=%q\t lit=%q\n", pos, p.fSet.Position(pos), tok, tok.String(), lit)
		if tok == token.EOF {
			if !beforeToken.CanEOF() {
				return nil, fmt.Errorf("%s can NOT be last", beforeToken.String())
			}
			break
		}
		if tok == token.COMMENT || tok == token.SEMICOLON {
			continue
		}

		if tok == token.LPAREN {
			lParenCount++
		}

		if tok == token.RPAREN {
			rParenCount++
		}

		var (
			symbol       = UNKNOWN
			parseToken   Token
			supported    bool
			parseTokenFn func(pos token.Pos, tok token.Token, lit string) (Token, error)
			err          error
		)

		if tok == token.IDENT {
			if beforeToken.Symbol() == FUNC {
				inputCustomFn, okInput := p.option.customFns[lit]
				buildInCustomFn, okBuildIn := _buildInCustomFn[lit]
				if !okInput && !okBuildIn {
					return nil, fmt.Errorf("unknown func, name=%s, pos=%v", lit, pos)
				}
				tokenFn := beforeToken.(*tokenFunc)
				tokenFn.lit = lit
				if okInput {
					tokenFn.value = inputCustomFn
					continue
				}
				if okBuildIn {
					tokenFn.value = buildInCustomFn
					continue
				}
			}

			// like `foo.`, `foo.bar.` , must be accessor
			if beforeToken.Symbol() == PERIOD {
				before := beforeToken.(*tokenIdent)
				withPeriodStr := fmt.Sprintf("%s%s", before.lit, lit)
				before.lit = withPeriodStr
				before.value = withPeriodStr
				continue
			}

			// parse bool
			if strings.ToUpper(lit) == "TRUE" || strings.ToUpper(lit) == "FALSE" {
				symbol = BOOL
				goto symbolStep
			}
		}

		// like `foo`, `bar` + `.`, must be accessor
		if tok == token.PERIOD {
			if beforeToken.Symbol() == IDENT {
				before := beforeToken.(*tokenIdent)
				withPeriodStr := fmt.Sprintf("%s.", before.lit)
				before.lit = withPeriodStr
				before.value = withPeriodStr
				continue
			}
		}

		// like `.1`, `.2` ... must be accessor and try to get data from array with array index (foo.bar.1)
		if tok == token.FLOAT && strings.LastIndex(lit, ".") == 0 {
			if beforeToken.Symbol() == IDENT {
				before := beforeToken.(*tokenIdent)
				withPeriodStr := fmt.Sprintf("%s%s", before.lit, lit)
				before.lit = withPeriodStr
				before.value = withPeriodStr
				continue
			}
		}

		if beforeToken.CanNext(&tokenNEGATE{}) == nil && tok == token.SUB {
			symbol = NEGATE
			goto symbolStep
		}

		symbol, supported = token2Symbol[tok]
		if !supported {
			return nil, fmt.Errorf("token2Symbol unsupported expr, token=%s\t lit=%s\t pos:%v", tok.String(), lit, p.fSet.Position(pos))
		}

	symbolStep:
		parseTokenFn, supported = symbol2Token[symbol]
		if !supported {
			return nil, fmt.Errorf("symbol2Token unsupported expr, token=%s\t lit=%s\t pos:%v", tok.String(), lit, p.fSet.Position(pos))
		}
		parseToken, err = parseTokenFn(pos, tok, lit)
		if err != nil {
			return nil, err
		}

		// check if current is valid
		err = beforeToken.CanNext(parseToken)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, parseToken)
		beforeToken = parseToken
	}

	if lParenCount != rParenCount {
		return nil, fmt.Errorf("check got %d left paren but %d right paren, should be equal", lParenCount, rParenCount)
	}

	return tokens, nil
}

func buildStage(stream *stream) (*stage, error) {

	if !stream.hasNext() {
		return nil, nil
	}

	stage, err := planSeparator(stream)
	if err != nil {
		return nil, err
	}
	reorderStages(stage)

	return stage, nil
}

//	During stage planning, stages of equal precedence are parsed such that they'll be evaluated in reverse order.
//
//	For commutative operators like "+" or "-", it's no big deal.
//	But for order-specific operators, it ruins the expected result.
func reorderStages(rootStage *stage) {

	if rootStage == nil {
		return
	}

	// traverse every rightStage until we find multiples in a row of the same precedence.
	var identicalPrecedences []*stage
	var currentStage, nextStage *stage
	var precedence, currentPrecedence int

	nextStage = rootStage
	precedence = symbolPrecedence(rootStage.symbol.Symbol())

	for nextStage != nil {

		currentStage = nextStage
		nextStage = currentStage.right

		// left depth first, since this entire method only looks for precedences down the right side of the tree
		if currentStage.left != nil {
			reorderStages(currentStage.left)
		}

		currentPrecedence = symbolPrecedence(currentStage.symbol.Symbol())

		if currentPrecedence == precedence {
			identicalPrecedences = append(identicalPrecedences, currentStage)
			continue
		}

		// precedence break.
		// See how many in a row we had, and reorder if there's more than one.
		if len(identicalPrecedences) > 1 {
			mirrorStageSubtree(identicalPrecedences)
		}

		identicalPrecedences = []*stage{currentStage}
		precedence = currentPrecedence
	}

	if len(identicalPrecedences) > 1 {
		mirrorStageSubtree(identicalPrecedences)
	}
}

//	Performs a "mirror" on a subtree of stages.
//
//	This mirror functionally inverts the order of execution for all members of the [stages] list.
//	That list is assumed to be a root-to-leaf (ordered) list of evaluation stages,
//	where each is a right-hand stage of the last.
func mirrorStageSubtree(stages []*stage) {

	var rootStage, inverseStage, carryStage, frontStage *stage

	stagesLength := len(stages)

	// reverse all right/left
	for _, frontStage = range stages {

		carryStage = frontStage.right
		frontStage.right = frontStage.left
		frontStage.left = carryStage
	}

	// end left swaps with root right
	rootStage = stages[0]
	frontStage = stages[stagesLength-1]

	carryStage = frontStage.left
	frontStage.left = rootStage.right
	rootStage.right = carryStage

	// for all non-root non-end stages, right is swapped with inverse stage right in list
	for i := 0; i < (stagesLength-2)/2+1; i++ {

		frontStage = stages[i+1]
		inverseStage = stages[stagesLength-i-1]

		carryStage = frontStage.right
		frontStage.right = inverseStage.right
		inverseStage.right = carryStage
	}

	// swap all other information with inverse stages
	for i := 0; i < stagesLength/2; i++ {
		frontStage = stages[i]
		inverseStage = stages[stagesLength-i-1]
		frontStage.swapWith(inverseStage)
	}
}

func doStage(stage *stage, parameters interface{}) (interface{}, error) {

	if stage == nil {
		return nil, nil
	}

	var left, right interface{}
	var err error

	if stage.left != nil {
		left, err = doStage(stage.left, parameters)
		if err != nil {
			return nil, err
		}
	}

	if stage.right != nil {
		right, err = doStage(stage.right, parameters)
		if err != nil {
			return nil, err
		}
	}

	if stage.symbol.LeftCheckFn() != nil {
		err = stage.symbol.LeftCheckFn()(left, right, parameters)
		if err != nil {
			return nil, err
		}
	}

	if stage.symbol.RightCheckFn() != nil {
		err = stage.symbol.RightCheckFn()(left, right, parameters)
		if err != nil {
			return nil, err
		}
	}

	return stage.symbol.SymbolFn()(left, right, parameters)
}

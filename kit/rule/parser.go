package rule

import (
	"context"
	"fmt"
	"go/scanner"
	"go/token"
	"log"
	"os"
	"strings"
	"time"
)

var (
	logger = log.New(os.Stdout, fmt.Sprintf("[DEBUG][pkg=rule][%s] ", time.Now().Format(time.StampMilli)), log.Lshortfile)
)

func init() {
	//logger.SetFlags(0)
	//logger.SetOutput(io.Discard)
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

func Do(ctx context.Context, expr string, params map[string]interface{}, options ...WithOption) (interface{}, error) {
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

func (p *Parser) Eval(params map[string]interface{}) (interface{}, error) {
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
		logger.Printf("pos_o=%d  pos=%s\t token=%q\t lit=%q\n", pos, p.fSet.Position(pos), tok.String(), lit)
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
			symbol       = Unkown
			parseToken   Token
			supported    bool
			parseTokenFn func(pos token.Pos, tok token.Token, lit string) (Token, error)
			err          error
		)

		if tok == token.IDENT {
			if beforeToken.Symbol() == Func {
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

			// parse bool
			if strings.ToUpper(lit) == "TRUE" || strings.ToUpper(lit) == "FALSE" {
				symbol = Bool
				goto symbolStep
			}
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

	return planSeparator(stream)
}

func buildStageFromTokens(tokens []Token) (*stage, error) {
	ret := new(stream)
	ret.tokens = tokens
	ret.len = len(tokens)

	return planSeparator(ret)
}

func doStage(stage *stage, parameters map[string]interface{}) (interface{}, error) {

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

package rule

import (
	"context"
	"fmt"
	"go/scanner"
	"go/token"
)

type parser struct {
	sc   scanner.Scanner
	fSet *token.FileSet
	file *token.File
}

func newParser(ctx context.Context, expr string) *parser {
	src := []byte(expr)

	var s scanner.Scanner
	fSet := token.NewFileSet()
	file := fSet.AddFile("", fSet.Base(), len(src))
	s.Init(file, src, nil /* no error handler */, scanner.ScanComments)

	return &parser{sc: s, fSet: fSet, file: file}
}

func (p *parser) read() ([]Token, error) {
	tokens := make([]Token, 0)
	for {
		pos, tok, lit := p.sc.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.COMMENT || tok == token.SEMICOLON {
			continue
		}
		symbol, supported := token2Symbol[tok]
		if !supported {
			return nil, fmt.Errorf("unsupported expr, token=%s\t lit=%s\t pos:%v", tok.String(), lit, p.fSet.Position(pos))
		}

		buildIn, supported := symbol2Token[symbol]
		if !supported {
			return nil, fmt.Errorf("unsupported expr, token=%s\t lit=%s\t pos:%v", tok.String(), lit, p.fSet.Position(pos))
		}
		in, err := buildIn(pos, tok, lit)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, in)
	}
	return tokens, nil
}

func (p *parser) stream() (ret *stream, err error) {
	ret = new(stream)
	ret.tokens, err = p.read()
	ret.len = len(ret.tokens)
	return ret, err
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

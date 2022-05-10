package rule

import (
	"context"
	"testing"
)

func Benchmark_SingleParse(bench *testing.B) {

	for i := 0; i < bench.N; i++ {
		New(context.Background(), "1")
	}
}

func Benchmark_Single(bench *testing.B) {

	parser, _ := New(context.Background(), "1")

	bench.ResetTimer()
	for i := 0; i < bench.N; i++ {
		parser.Eval(nil)
	}
}

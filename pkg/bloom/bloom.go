package bloom

import (
	"context"
)

type (
	// Filter is a bloom filter
	Filter struct {

		// todo counter
		//total int64
		//hit   int64
		//miss  int64

		Provider
	}

	Provider interface {
		Add(ctx context.Context, data []byte) error
		Exists(ctx context.Context, data []byte) (bool, error)
	}
)

func NewWithProvider(provider Provider) *Filter {
	return &Filter{Provider: provider}
}

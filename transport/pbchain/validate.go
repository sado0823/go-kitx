package pbchain

import (
	"context"

	"github.com/sado0823/go-kitx/errorx"
)

type validator interface {
	Validate() error
}

func Validator() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req interface{}) (resp interface{}, err error) {
			if v, ok := req.(validator); ok {
				if err = v.Validate(); err != nil {
					return nil, errorx.BadRequest("VALIDATOR", err.Error()).WithCause(err)
				}
			}
			return next(ctx, req)
		}
	}
}

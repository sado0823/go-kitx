package middleware

import "context"

type (
	Handler func(ctx context.Context, req interface{}) (resp interface{}, err error)

	// Middleware is a transport http/grpc middleware
	Middleware func(next Handler) Handler
)

func Chain(m ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(m) - 1; i >= 0; i-- {
			next = m[i](next)
		}
		return next
	}
}

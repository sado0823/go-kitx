package pbchain

import (
	"context"
	"runtime"

	"github.com/sado0823/go-kitx/errorx"
	"github.com/sado0823/go-kitx/kit/log"
)

func Recovery() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			defer func() {
				if rerr := recover(); rerr != nil {
					buf := make([]byte, 64<<10)
					n := runtime.Stack(buf, false)
					buf = buf[:n]
					log.Context(ctx).Errorf("panic recover %v: %+v\n%s\n", rerr, req, buf)

					err = errorx.InternalServer("panic", "500")
				}
			}()
			return next(ctx, req)
		}
	}
}

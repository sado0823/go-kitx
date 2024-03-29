package pbchain

import (
	"context"
	"fmt"
	"time"

	"github.com/sado0823/go-kitx/errorx"
	"github.com/sado0823/go-kitx/kit/log"
	"github.com/sado0823/go-kitx/transport"
)

func LoggingServer() Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, req interface{}) (resp interface{}, err error) {
			var (
				code         int32
				reason       string
				kind         string
				operation    string
				pathTemplate string
			)
			startTime := time.Now()
			if info, ok := transport.FromServerContext(ctx); ok {
				kind = info.Kind().String()
				operation = info.Operation()
				pathTemplate = info.PathTemplate()
			}
			resp, err = next(ctx, req)
			if se := errorx.FromError(err); se != nil {
				code = se.Code
				reason = se.Reason
			}
			level, stack := extractError(err)
			log.Context(ctx).Log(level,
				"kind", "server",
				"component", kind,
				"operation", operation,
				"path_template", pathTemplate,
				"args", extractArgs(req),
				"code", code,
				"reason", reason,
				"stack", stack,
				"latency", fmt.Sprintf("%f", time.Since(startTime).Seconds()),
			)
			return
		}
	}
}

func LoggingClient() Middleware {
	return func(handler Handler) Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var (
				code         int32
				reason       string
				kind         string
				operation    string
				pathTemplate string
			)
			startTime := time.Now()
			if info, ok := transport.FromClientContext(ctx); ok {
				kind = info.Kind().String()
				operation = info.Operation()
				pathTemplate = info.PathTemplate()
			}
			reply, err = handler(ctx, req)
			if se := errorx.FromError(err); se != nil {
				code = se.Code
				reason = se.Reason
			}
			level, stack := extractError(err)
			log.Context(ctx).Log(level,
				"kind", "client",
				"component", kind,
				"operation", operation,
				"path_template", pathTemplate,
				"args", extractArgs(req),
				"code", code,
				"reason", reason,
				"stack", stack,
				"latency", fmt.Sprintf("%f", time.Since(startTime).Seconds()),
			)
			return
		}
	}
}

// extractArgs returns the string of the req
func extractArgs(req interface{}) string {
	if stringer, ok := req.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%+v", req)
}

// extractError returns the string of the error
func extractError(err error) (log.Level, string) {
	if err != nil {
		return log.LevelError, fmt.Sprintf("%+v", err)
	}
	return log.LevelInfo, ""
}

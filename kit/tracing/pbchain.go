package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	"github.com/sado0823/go-kitx/kit/log"
	"github.com/sado0823/go-kitx/transport"
	"github.com/sado0823/go-kitx/transport/pbchain"
)

// Server returns a new server middleware for OpenTelemetry.
func Server() pbchain.Middleware {
	tracer := NewTracer(trace.SpanKindServer, WithTracerName("go-kitx"))
	return func(handler pbchain.Handler) pbchain.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				var span trace.Span
				ctx, span = tracer.Start(ctx, tr.PathTemplate(), tr.RequestHeader())
				SetServerSpan(ctx, span, req)
				defer func() { tracer.End(ctx, span, reply, err) }()
			}
			return handler(ctx, req)
		}
	}
}

// Client returns a new client middleware for OpenTelemetry.
func Client() pbchain.Middleware {
	tracer := NewTracer(trace.SpanKindClient, WithTracerName("go-kitx"))
	return func(handler pbchain.Handler) pbchain.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromClientContext(ctx); ok {
				var span trace.Span
				ctx, span = tracer.Start(ctx, tr.Operation(), tr.RequestHeader())
				SetClientSpan(ctx, span, req)
				defer func() { tracer.End(ctx, span, reply, err) }()
			}
			return handler(ctx, req)
		}
	}
}

// TraceID returns a traceid valuer.
func TraceID() log.Valuer {
	return func(ctx context.Context) interface{} {
		if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
			return span.TraceID().String()
		}
		return ""
	}
}

// SpanID returns a spanid valuer.
func SpanID() log.Valuer {
	return func(ctx context.Context) interface{} {
		if span := trace.SpanContextFromContext(ctx); span.HasSpanID() {
			return span.SpanID().String()
		}
		return ""
	}
}

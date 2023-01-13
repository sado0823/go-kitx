package sqlx

import (
	"context"
	"database/sql"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	traceName = "kitx"
	spanSql   = "sql"
)

var sqlSpanAttributeKey = attribute.Key("sql.method")

func startSpan(ctx context.Context, method string) (startCtx context.Context, span oteltrace.Span) {
	startCtx, span = otel.Tracer(traceName).
		Start(ctx, spanSql, oteltrace.WithSpanKind(oteltrace.SpanKindClient))
	span.SetAttributes(sqlSpanAttributeKey.String(method))

	return startCtx, span
}

func endSpan(span oteltrace.Span, err error) {
	defer span.End()

	if err == nil || errors.Is(err, sql.ErrNoRows) {
		span.SetStatus(codes.Ok, "")
		return
	}

	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}

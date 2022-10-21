package log

import (
	"context"
	"log"
)

type Logger interface {
	Log(level Level, kvs ...interface{}) error
}

var defaultLogger = WithFields(NewStd(log.Writer()), "ts", DefaultTimestamp, "caller", DefaultCaller)

type logger struct {
	ctx       context.Context
	internal  Logger
	prefix    []interface{}
	hasValuer bool
}

func (l *logger) Log(level Level, kvs ...interface{}) error {
	fullKvs := make([]interface{}, 0, len(kvs)+len(l.prefix))
	fullKvs = append(fullKvs, l.prefix...)
	if l.hasValuer {
		bindValues(l.ctx, fullKvs)
	}
	fullKvs = append(fullKvs, kvs...)
	return l.internal.Log(level, fullKvs...)
}

func RemoveFields(l Logger, keys ...interface{}) Logger {
	from, ok := l.(*logger)
	if !ok {
		return l
	}

	lenKvs := len(from.prefix)
	kvs := from.prefix
	if (lenKvs & 1) == 1 {
		kvs = append(kvs, "log key value unpaired")
	}

	newKvs := make([]interface{}, 0, len(kvs))
	for i := 0; i < lenKvs; i += 2 {
		for _, key := range keys {
			if kvs[i] == key {
				continue
			}
			newKvs = append(newKvs, kvs[i], kvs[i+1])
		}
	}

	return &logger{ctx: from.ctx, internal: from.internal, prefix: newKvs, hasValuer: containValuer(newKvs)}
}

// WithFields add new fields to the logger
func WithFields(l Logger, kvs ...interface{}) Logger {
	from, ok := l.(*logger)
	if !ok {
		return &logger{ctx: context.Background(), internal: l, prefix: kvs, hasValuer: containValuer(kvs)}
	}

	fullKvs := make([]interface{}, 0, len(kvs)+len(from.prefix))
	fullKvs = append(fullKvs, from.prefix...)
	fullKvs = append(fullKvs, kvs...)
	return &logger{ctx: from.ctx, internal: from.internal, prefix: fullKvs, hasValuer: containValuer(fullKvs)}
}

// WithContext return a shadow copy of logger with a new context
func WithContext(ctx context.Context, l Logger) Logger {
	from, ok := l.(*logger)
	if !ok {
		return &logger{ctx: context.Background(), internal: l}
	}

	return &logger{ctx: ctx, internal: from.internal, prefix: from.prefix, hasValuer: from.hasValuer}
}

package log

import (
	"context"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Valuer func(ctx context.Context) interface{}

var (
	DefaultCaller    = Caller(4)
	DefaultTimestamp = Timestamp(time.RFC3339)
)

func Timestamp(layout string) Valuer {
	return func(ctx context.Context) interface{} {
		return time.Now().Format(layout)
	}
}

func Caller(depth int) Valuer {
	return func(ctx context.Context) interface{} {
		_, file, line, _ := runtime.Caller(depth)
		idx := strings.LastIndexByte(file, '/')
		if idx == -1 {
			return file[idx+1:] + ":" + strconv.Itoa(line)
		}
		idx = strings.LastIndexByte(file[:idx], '/')
		return file[idx+1:] + ":" + strconv.Itoa(line)
	}
}

func bindValues(ctx context.Context, kvs []interface{}) {
	for i := 1; i < len(kvs); i += 2 {
		if fn, ok := kvs[i].(Valuer); ok {
			kvs[i] = fn(ctx)
		}
	}
}

func containValuer(kvs []interface{}) bool {
	for i := 1; i < len(kvs); i += 2 {
		if _, ok := kvs[i].(Valuer); ok {
			return true
		}
	}
	return false
}

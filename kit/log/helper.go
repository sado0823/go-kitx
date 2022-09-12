package log

import (
	"context"
	"fmt"
	"os"
)

type (
	HelperOption func(helper *Helper)

	Helper struct {
		logger Logger
		msgKey string
	}
)

var DefaultMessageKey = "msg"

func WithHelperMessageKey(msgKey string) HelperOption {
	return func(helper *Helper) {
		helper.msgKey = msgKey
	}
}

func NewHelper(logger Logger, options ...HelperOption) *Helper {
	ops := &Helper{logger: logger, msgKey: DefaultMessageKey}
	for _, option := range options {
		option(ops)
	}

	return ops
}

func (h *Helper) WithContext(ctx context.Context) *Helper {
	return &Helper{
		msgKey: h.msgKey,
		logger: WithContext(ctx, h.logger),
	}
}

func (h *Helper) Log(level Level, keyvals ...interface{}) {
	_ = h.logger.Log(level, keyvals...)
}

func (h *Helper) Debug(a ...interface{}) {
	_ = h.logger.Log(LevelDebug, h.msgKey, fmt.Sprint(a...))
}

func (h *Helper) Debugf(format string, a ...interface{}) {
	_ = h.logger.Log(LevelDebug, h.msgKey, fmt.Sprintf(format, a...))
}

func (h *Helper) Debugw(keyvals ...interface{}) {
	_ = h.logger.Log(LevelDebug, keyvals...)
}

func (h *Helper) Info(a ...interface{}) {
	_ = h.logger.Log(LevelInfo, h.msgKey, fmt.Sprint(a...))
}

func (h *Helper) Infof(format string, a ...interface{}) {
	_ = h.logger.Log(LevelInfo, h.msgKey, fmt.Sprintf(format, a...))
}

func (h *Helper) Infow(keyvals ...interface{}) {
	_ = h.logger.Log(LevelInfo, keyvals...)
}

func (h *Helper) Warn(a ...interface{}) {
	_ = h.logger.Log(LevelWarn, h.msgKey, fmt.Sprint(a...))
}

func (h *Helper) Warnf(format string, a ...interface{}) {
	_ = h.logger.Log(LevelWarn, h.msgKey, fmt.Sprintf(format, a...))
}

func (h *Helper) Warnw(keyvals ...interface{}) {
	_ = h.logger.Log(LevelWarn, keyvals...)
}

func (h *Helper) Error(a ...interface{}) {
	_ = h.logger.Log(LevelError, h.msgKey, fmt.Sprint(a...))
}

func (h *Helper) Errorf(format string, a ...interface{}) {
	_ = h.logger.Log(LevelError, h.msgKey, fmt.Sprintf(format, a...))
}

func (h *Helper) Errorw(keyvals ...interface{}) {
	_ = h.logger.Log(LevelError, keyvals...)
}

func (h *Helper) Fatal(a ...interface{}) {
	_ = h.logger.Log(LevelFatal, h.msgKey, fmt.Sprint(a...))
	os.Exit(1)
}

func (h *Helper) Fatalf(format string, a ...interface{}) {
	_ = h.logger.Log(LevelFatal, h.msgKey, fmt.Sprintf(format, a...))
	os.Exit(1)
}

func (h *Helper) Fatalw(keyvals ...interface{}) {
	_ = h.logger.Log(LevelFatal, keyvals...)
	os.Exit(1)
}
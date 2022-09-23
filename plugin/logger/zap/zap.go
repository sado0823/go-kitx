package zap

import (
	"github.com/sado0823/go-kitx/kit/log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_ log.Logger = (*Logger)(nil)

	levelTransfer = map[log.Level]zapcore.Level{
		log.LevelDebug: zap.DebugLevel,
		log.LevelInfo:  zap.InfoLevel,
		log.LevelWarn:  zap.WarnLevel,
		log.LevelError: zap.ErrorLevel,
		log.LevelFatal: zap.FatalLevel,
	}
)

type Logger struct {
	internal *zap.Logger
}

func New(logger *zap.Logger) *Logger {
	return &Logger{internal: logger}
}

func (l *Logger) Log(level log.Level, kvs ...interface{}) error {
	if len(kvs) == 0 {
		return nil
	}

	if len(kvs)%2 != 0 {
		kvs = append(kvs, "")
	}

	zapLevel, ok := levelTransfer[level]
	if !ok {
		zapLevel = zapcore.InfoLevel
	}

	var (
		fields []zap.Field
		msg    string
	)
	for i := 0; i < len(kvs); i += 2 {
		key, ok := kvs[i].(string)
		if !ok {
			continue
		}
		if key == "msg" {
			msg, _ = kvs[i+1].(string)
			continue
		}
		fields = append(fields, zap.Any(key, kvs[i+1]))
	}

	l.internal.Log(zapLevel, msg, fields...)

	return nil
}

func (l *Logger) Sync() error {
	return l.internal.Sync()
}

func (l *Logger) Close() error {
	return l.Sync()
}

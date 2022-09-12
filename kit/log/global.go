package log

import (
	"context"
	"fmt"
	"os"
	"sync"
)

var globaler = &global{}

type global struct {
	lock sync.Mutex
	Logger
}

func init() {
	globaler.set(defaultLogger)
}

func (g *global) set(l Logger) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.Logger = l
}

// SetGlobal should be called before init App
func SetGlobal(l Logger) {
	globaler.set(l)
}

func GetGlobal() Logger {
	return globaler.Logger
}

func Log(level Level, keyvals ...interface{}) {
	_ = globaler.Log(level, keyvals...)
}

func Context(ctx context.Context) *Helper {
	return NewHelper(WithContext(ctx, globaler.Logger))
}

func Debug(a ...interface{}) {
	_ = globaler.Log(LevelDebug, DefaultMessageKey, fmt.Sprint(a...))
}

func Debugf(format string, a ...interface{}) {
	_ = globaler.Log(LevelDebug, DefaultMessageKey, fmt.Sprintf(format, a...))
}

func Debugw(keyvals ...interface{}) {
	_ = globaler.Log(LevelDebug, keyvals...)
}

func Info(a ...interface{}) {
	_ = globaler.Log(LevelInfo, DefaultMessageKey, fmt.Sprint(a...))
}

func Infof(format string, a ...interface{}) {
	_ = globaler.Log(LevelInfo, DefaultMessageKey, fmt.Sprintf(format, a...))
}

func Infow(keyvals ...interface{}) {
	_ = globaler.Log(LevelInfo, keyvals...)
}

func Warn(a ...interface{}) {
	_ = globaler.Log(LevelWarn, DefaultMessageKey, fmt.Sprint(a...))
}

func Warnf(format string, a ...interface{}) {
	_ = globaler.Log(LevelWarn, DefaultMessageKey, fmt.Sprintf(format, a...))
}

func Warnw(keyvals ...interface{}) {
	_ = globaler.Log(LevelWarn, keyvals...)
}

func Error(a ...interface{}) {
	_ = globaler.Log(LevelError, DefaultMessageKey, fmt.Sprint(a...))
}

func Errorf(format string, a ...interface{}) {
	_ = globaler.Log(LevelError, DefaultMessageKey, fmt.Sprintf(format, a...))
}

func Errorw(keyvals ...interface{}) {
	_ = globaler.Log(LevelError, keyvals...)
}

func Fatal(a ...interface{}) {
	_ = globaler.Log(LevelFatal, DefaultMessageKey, fmt.Sprint(a...))
	os.Exit(1)
}

func Fatalf(format string, a ...interface{}) {
	_ = globaler.Log(LevelFatal, DefaultMessageKey, fmt.Sprintf(format, a...))
	os.Exit(1)
}

func Fatalw(keyvals ...interface{}) {
	_ = globaler.Log(LevelFatal, keyvals...)
	os.Exit(1)
}

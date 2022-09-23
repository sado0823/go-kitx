package logrus

import (
	"github.com/sado0823/go-kitx/kit/log"

	"github.com/sirupsen/logrus"
)

var (
	_ log.Logger = (*Logger)(nil)

	levelTransfer = map[log.Level]logrus.Level{
		log.LevelDebug: logrus.DebugLevel,
		log.LevelInfo:  logrus.InfoLevel,
		log.LevelWarn:  logrus.WarnLevel,
		log.LevelError: logrus.ErrorLevel,
		log.LevelFatal: logrus.FatalLevel,
	}
)

type Logger struct {
	internal *logrus.Logger
}

func New(logger *logrus.Logger) log.Logger {
	return &Logger{
		internal: logger,
	}
}

func (l *Logger) Log(level log.Level, kvs ...interface{}) error {
	if len(kvs) == 0 {
		return nil
	}

	logrusLevel, ok := levelTransfer[level]
	if !ok {
		logrusLevel = logrus.InfoLevel
	}

	if logrusLevel > l.internal.Level {
		return nil
	}

	if len(kvs)%2 != 0 {
		kvs = append(kvs, "")
	}

	var (
		fields logrus.Fields = make(map[string]interface{})
		msg    string
	)

	for i := 0; i < len(kvs); i += 2 {
		key, ok := kvs[i].(string)
		if !ok {
			continue
		}
		if key == logrus.FieldKeyMsg {
			msg, _ = kvs[i+1].(string)
			continue
		}
		fields[key] = kvs[i+1]
	}

	if len(fields) > 0 {
		l.internal.WithFields(fields).Log(logrusLevel, msg)
	} else {
		l.internal.Log(logrusLevel, msg)
	}

	return nil
}

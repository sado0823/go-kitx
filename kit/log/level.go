package log

import "strings"

type Level int8

const (
	LevelDebug Level = iota - 1
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

var (
	level2Str = map[Level]string{
		LevelDebug: "DEBUG",
		LevelInfo:  "INFO",
		LevelWarn:  "WARN",
		LevelError: "ERROR",
		LevelFatal: "FATAL",
	}
	str2Level = map[string]Level{
		"DEBUG": LevelDebug,
		"INFO":  LevelInfo,
		"WARN":  LevelWarn,
		"ERROR": LevelError,
		"FATAL": LevelFatal,
	}
)

func (l Level) String() string {
	s, ok := level2Str[l]
	if !ok {
		return ""
	}
	return s
}

func ToLevel(s string) Level {
	upper := strings.ToUpper(s)
	level, ok := str2Level[upper]
	if !ok {
		return LevelInfo
	}
	return level
}

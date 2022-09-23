package log

import "fmt"

const (
	colorFgBlack = iota + 30
	colorFgRed
	colorFgGreen
	colorFgYellow
	colorFgBlue
	colorFgMagenta
	colorFgCyan
	colorFgWhite
)

var level2Color = map[Level]int{
	LevelDebug: colorFgWhite,
	LevelInfo:  colorFgGreen,
	LevelWarn:  colorFgYellow,
	LevelError: colorFgRed,
	LevelFatal: colorFgMagenta,
}

func colorLevel(level Level) string {
	color, ok := level2Color[level]
	if !ok {
		color = colorFgBlack
	}

	return fmt.Sprintf("\x1b[%dm%s\x1b[0m ", color, level.String())
}

func withColor(level Level, origin interface{}) string {
	color, ok := level2Color[level]
	if !ok {
		color = colorFgBlack
	}

	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", color, origin)
}

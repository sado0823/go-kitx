package log

import (
	color "github.com/sado0823/go-kitx/kit/log/internal"
)

func colorLevel(levelStr string) string {
	var colour color.Color
	switch levelStr {
	case LevelDebug.String():
		colour = color.BgBlue
	case LevelInfo.String():
		colour = color.BgGreen
	case LevelWarn.String():
		colour = color.BgCyan
	case LevelError.String():
		colour = color.BgRed
	case LevelFatal.String():
		colour = color.BgMagenta
	default:
		colour = color.NoColor
	}

	if colour == color.NoColor {
		return levelStr
	}

	return color.WithColorPadding(levelStr, colour)
}

package logging

import (
	"log/slog"

	"fmt"
	"runtime"
	"strings"

	"github.com/fatih/color"
)

// colorizeByLevel returns colored versions of level and message strings
func colorizeByLevel(level slog.Level, levelString, message string) (string, string) {
	switch level {
	case slog.LevelDebug:
		return color.BlueString(levelString), color.BlueString(message)
	case slog.LevelInfo:
		return color.GreenString(levelString), color.GreenString(message)
	case slog.LevelWarn:
		return color.YellowString(levelString), color.YellowString(message)
	case slog.LevelError:
		return color.RedString(levelString), color.RedString(message)
	default:
		return levelString, message
	}
}

// colorizeString returns a colored string based on log level
func colorizeString(level slog.Level, text string) string {
	switch level {
	case slog.LevelDebug:
		return color.BlueString(text)
	case slog.LevelInfo:
		return color.GreenString(text)
	case slog.LevelWarn:
		return color.YellowString(text)
	case slog.LevelError:
		return color.RedString(text)
	default:
		return text
	}
}

// extractSourceInfo extracts file and line information from a record
func extractSourceInfo(record slog.Record) string {
	if record.PC == 0 {
		return ""
	}

	frame := runtime.CallersFrames([]uintptr{record.PC})
	if frame == nil {
		return ""
	}

	f, _ := frame.Next()
	if f.File == "" {
		return ""
	}

	file := f.File
	if idx := strings.LastIndex(file, "/"); idx != -1 {
		file = file[idx+1:]
	}

	return fmt.Sprintf("%s:%d", file, f.Line)
}

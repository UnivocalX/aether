package logging

import (
	"log/slog"
	"os"
	"strings"
)

type Mode string

const (
	BaseMode   Mode = "base"
	CLIMode    Mode = "cli"
	ServerMode Mode = "server"
)

type Log struct {
	mode    Mode
	level   slog.Level
	colored bool
}

type LogOption func(*Log)

func (l *Log) Apply() {
	var handler slog.Handler
	switch l.mode {
	case CLIMode:
		handler = NewCliHandler(l.level)
	case ServerMode:
		handler = NewJSONHandler(l.level, l.colored)
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: l.level})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func (l *Log) SetMode(mode Mode) {
	l.mode = mode
}

func (l *Log) SetLevel(level string) {
	l.level = parseLogLevel(level)
}

func (l *Log) EnableColor() {
	l.colored = true
}

func NewLog(opts ...LogOption) *Log {
	l := &Log{
		mode:    BaseMode,
		level:   slog.LevelInfo,
		colored: false,
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// WithMode sets the logging mode (base, cli, or server)
func WithMode(mode Mode) LogOption {
	return func(l *Log) {
		l.mode = mode
	}
}

// WithLevelString sets the log level from a string
func WithLevelString(levelStr string) LogOption {
	return func(l *Log) {
		l.level = parseLogLevel(levelStr)
	}
}

// parseLogLevel converts a string log level to slog.Level
func parseLogLevel(levelStr string) slog.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

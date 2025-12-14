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

type Logging struct {
	mode        Mode
	level       slog.Level
	coloredJSON bool
}

type LoggingOption func(*Logging)

func (l *Logging) Apply() {
	var handler slog.Handler
	switch l.mode {
	case CLIMode:
		handler = NewCliHandler(l.level)
	case ServerMode:
		handler = NewJSONHandler(l.level, l.coloredJSON)
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: l.level})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func (l *Logging) SetMode(mode Mode) {
	l.mode = mode
}

func (l *Logging) SetLevel(level string) {
	l.level = parseLogLevel(level)
}

func (l *Logging) SetColoredJSON(enabled bool) {
	l.coloredJSON = enabled
}

func NewLogging(opts ...LoggingOption) *Logging {
	l := &Logging{
		mode:        BaseMode,
		level:       slog.LevelInfo,
		coloredJSON: false,
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// WithMode sets the logging mode (base, cli, or server)
func WithMode(mode Mode) LoggingOption {
	return func(l *Logging) {
		l.mode = mode
	}
}

// WithLevelString sets the log level from a string
func WithLevelString(levelStr string) LoggingOption {
	return func(l *Logging) {
		l.level = parseLogLevel(levelStr)
	}
}

// WithColoredJSON enables colored JSON output in server mode
func WithColoredJSON(enabled bool) LoggingOption {
	return func(l *Logging) {
		l.coloredJSON = enabled
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
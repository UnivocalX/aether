package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Config holds configuration for service logging
type Config struct {
	Prod     bool // true for production, false for development
	AppName  string
	LogFile  string // Optional: custom log file path
	LogLevel slog.Level
}

// New creates a new simple logger that outputs only messages with colors
func New(w io.Writer, level slog.Level) *slog.Logger {
	if w == nil {
		w = os.Stderr
	}

	handler := &simpleHandler{
		handler: slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: level,
		}),
	}

	return slog.New(handler)
}

// NewService creates a new service logger with the specified configuration
func NewService(config *Config) (*slog.Logger, error) {
	// Set default log file if not provided
	logFile := config.LogFile
	if logFile == "" {
		logFile = getDefaultLogPath(config.AppName)
	}

	// Ensure log directory exists
	if err := ensureLogDir(filepath.Dir(logFile)); err != nil {
		return nil, err
	}

	// Open log file
	file, err := createLogFile(logFile)
	if err != nil {
		return nil, err
	}

	// Create handlers based on prod flag
	var consoleHandler slog.Handler
	var fileHandler slog.Handler

	if config.Prod {
		// Production mode - no colors, structured output
		consoleHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:       config.LogLevel,
			AddSource:   true,
			ReplaceAttr: replaceProdAttributes,
		})
		// JSON for file in production
		fileHandler = slog.NewJSONHandler(file, &slog.HandlerOptions{
			Level:       config.LogLevel,
			AddSource:   true,
			ReplaceAttr: replaceProdAttributes,
		})
	} else {
		// Development mode - with colors and friendly formatting
		consoleHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:       config.LogLevel,
			AddSource:   true,
			ReplaceAttr: replaceDevAttributes,
		})
		// JSON for file in development
		fileHandler = slog.NewJSONHandler(file, &slog.HandlerOptions{
			Level:       config.LogLevel,
			AddSource:   true,
			ReplaceAttr: replaceProdAttributes,
		})
	}

	// Create multi-handler
	multiHandler := newMultiHandler(consoleHandler, fileHandler)

	return slog.New(multiHandler), nil
}

// LevelFromString converts a string to slog.Level
// Supported levels: "debug", "info", "warn", "error"
// Returns LevelInfo for unknown values
func LevelFromString(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "err", "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // default to info
	}
}

// Internal helper functions

func ensureLogDir(logDir string) error {
	return os.MkdirAll(logDir, 0755)
}

func getDefaultLogPath(appName string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), appName, "logs", appName+".log")
	}
	return filepath.Join("/var/log", appName, appName+".log")
}

func createLogFile(path string) (*os.File, error) {
	if err := ensureLogDir(filepath.Dir(path)); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

// Color attributes for terminal output
const (
	colorRed     = 31
	colorGreen   = 32
	colorYellow  = 33
	colorBlue    = 34
	colorMagenta = 35
	colorCyan    = 36
	colorGray    = 37
	colorWhite   = 97
)

func colorize(colorCode int, text string) string {
	return "\033[" + strconv.Itoa(colorCode) + "m" + text + "\033[0m"
}

func levelColor(level slog.Level) int {
	switch {
	case level >= slog.LevelError:
		return colorRed
	case level >= slog.LevelWarn:
		return colorYellow
	case level >= slog.LevelInfo:
		return colorGreen
	default:
		return colorBlue // Use blue for debug
	}
}

// formatAttr formats a single attribute with improved formatting
func formatAttr(attr slog.Attr) string {
	key := attr.Key
	value := attr.Value.Any()

	// Handle different value types for better formatting
	switch v := value.(type) {
	case nil:
		return key
	case string:
		if v == "" {
			return key
		}
		return fmt.Sprintf("%s=%q", key, v)
	case error:
		if v == nil {
			return key
		}
		return fmt.Sprintf("%s=%q", key, v.Error())
	case fmt.Stringer:
		if v == nil {
			return key
		}
		return fmt.Sprintf("%s=%q", key, v.String())
	default:
		// For basic types, use simple formatting
		return fmt.Sprintf("%s=%v", key, value)
	}
}

// Simple Handler implementation - shows colored message with contextual attributes
type simpleHandler struct {
	handler slog.Handler
}

func (h *simpleHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *simpleHandler) Handle(ctx context.Context, r slog.Record) error {
	// Build the log line with colored message and attributes
	var logLine strings.Builder

	// Get color for this log level
	color := levelColor(r.Level)

	// Add colored message only (no level indicator)
	coloredMessage := colorize(color, r.Message)
	logLine.WriteString(coloredMessage)

	// Collect and format contextual attributes
	attrs := make([]string, 0)
	r.Attrs(func(attr slog.Attr) bool {
		if attr.Key != "" && attr.Key != slog.LevelKey && attr.Key != slog.TimeKey {
			formattedAttr := formatAttr(attr)
			coloredAttr := colorize(color, formattedAttr)
			attrs = append(attrs, coloredAttr)
		}
		return true
	})

	// Add contextual attributes if any, separated by commas
	if len(attrs) > 0 {
		logLine.WriteString(" [")
		logLine.WriteString(strings.Join(attrs, ", "))
		logLine.WriteString("]")
	}

	logLine.WriteString("\n")

	if _, err := io.WriteString(os.Stderr, logLine.String()); err != nil {
		return err
	}
	return nil
}

func (h *simpleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &simpleHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *simpleHandler) WithGroup(name string) slog.Handler {
	return &simpleHandler{handler: h.handler.WithGroup(name)}
}

// MultiHandler handles logging to multiple destinations
type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers ...slog.Handler) *multiHandler {
	return &multiHandler{handlers: handlers}
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			if err := handler.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return newMultiHandler(handlers...)
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return newMultiHandler(handlers...)
}

// Attribute replacement functions
func replaceDevAttributes(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		levelString := level.String()

		color := levelColor(level)
		coloredLevel := colorize(color, levelString)

		return slog.Attr{
			Key:   a.Key,
			Value: slog.StringValue(coloredLevel),
		}
	}

	if a.Key == slog.TimeKey && len(groups) == 0 {
		if t, ok := a.Value.Any().(time.Time); ok {
			return slog.Attr{
				Key:   a.Key,
				Value: slog.StringValue(t.Format("15:04:05.000")),
			}
		}
	}

	return a
}

func replaceProdAttributes(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey && len(groups) == 0 {
		if t, ok := a.Value.Any().(time.Time); ok {
			return slog.Attr{
				Key:   a.Key,
				Value: slog.StringValue(t.Format(time.RFC3339)),
			}
		}
	}
	return a
}

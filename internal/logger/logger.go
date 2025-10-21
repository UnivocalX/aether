package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
)

type Options struct {
	AddSource  bool
	Production bool
	Level      string
}

var log *slog.Logger

// Custom handler for development with colors
type colorfulHandler struct {
	handler slog.Handler
	level   slog.Level
}

func (h *colorfulHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *colorfulHandler) Handle(ctx context.Context, r slog.Record) error {
	// Define colors for different log levels
	var msgColor *color.Color
	switch r.Level {
	case slog.LevelDebug:
		msgColor = color.New(color.FgHiCyan)
	case slog.LevelInfo:
		msgColor = color.New(color.FgHiGreen)
	case slog.LevelWarn:
		msgColor = color.New(color.FgHiYellow)
	case slog.LevelError:
		msgColor = color.New(color.FgHiRed)
	default:
		msgColor = color.New(color.FgWhite)
	}

	// Print the message with appropriate color
	msgColor.Println(r.Message)

	// Print attributes on new lines with indentation
	if r.NumAttrs() > 0 {
		r.Attrs(func(attr slog.Attr) bool {
			key := color.HiCyanString("  " + attr.Key + ":")
			value := formatValue(attr.Value)
			fmt.Printf("%s %v\n", key, value)
			return true
		})
		// Add extra line break when contextual arguments are provided
		fmt.Println()
	}

	return nil
}

func formatValue(value slog.Value) interface{} {
	// All values use the same white color
	switch value.Kind() {
	case slog.KindString:
		return color.WhiteString("%q", value.String())
	case slog.KindInt64:
		return color.WhiteString("%d", value.Int64())
	case slog.KindBool:
		return color.WhiteString("%t", value.Bool())
	case slog.KindFloat64:
		return color.WhiteString("%.2f", value.Float64())
	case slog.KindTime:
		return color.WhiteString(value.Time().Format("15:04:05"))
	default:
		return color.WhiteString("%v", value.Any())
	}
}

func (h *colorfulHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &colorfulHandler{
		handler: h.handler.WithAttrs(attrs),
		level:   h.level,
	}
}

func (h *colorfulHandler) WithGroup(name string) slog.Handler {
	return &colorfulHandler{
		handler: h.handler.WithGroup(name),
		level:   h.level,
	}
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Init(cfg Options) {
	level := parseLevel(cfg.Level)

	var handler slog.Handler
	if cfg.Production {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					if t, ok := a.Value.Any().(time.Time); ok {
						a.Value = slog.StringValue(t.Format(time.RFC3339))
					}
				}
				return a
			},
		})
	} else {
		// Use our custom colorful handler for development
		baseHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     level,
			AddSource: cfg.AddSource,
		})
		handler = &colorfulHandler{
			handler: baseHandler,
			level:   level,
		}
	}

	log = slog.New(handler)
	slog.SetDefault(log)
}

func InitDefault() {
	Init(Options{
		AddSource:  false,
		Production: false,
		Level:      "info",
	})
}

// Convenience logging functions using structured attributes
func Info(msg string, args ...any)  { log.Info(msg, args...) }
func Error(msg string, args ...any) { log.Error(msg, args...) }
func Debug(msg string, args ...any) { log.Debug(msg, args...) }
func Warn(msg string, args ...any)  { log.Warn(msg, args...) }

// Contextual logging
func With(args ...any) *slog.Logger {
	return log.With(args...)
}

func ErrorErr(err error, msg string, args ...any) {
	args = append(args, slog.String("error", err.Error()))
	log.Error(msg, args...)
}

func DebugWithCaller(msg string, args ...any) {
	if pc, file, line, ok := runtime.Caller(1); ok {
		fn := runtime.FuncForPC(pc)
		args = append(args,
			slog.String("caller_func", fn.Name()),
			slog.String("caller_file", file),
			slog.Int("caller_line", line),
		)
	}
	log.Debug(msg, args...)
}

func Get() *slog.Logger {
	return log
}

func Sync() error {
	return nil
}
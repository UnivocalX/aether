package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

type Options struct {
	AddSource  bool
	Production bool
	Level      string
}

func (opt *Options) GetLevelValue() (slog.Level, error) {
	level := strings.ToLower(opt.Level)
	switch level {
	case "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("invalid log level %q, must be one of: debug, info, warn, error", level)
	}
}

// colorfulHandler implements slog.Handler with inline colorful formatting
type colorfulHandler struct {
	handler slog.Handler
	level   slog.Level
}

func (h *colorfulHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *colorfulHandler) Handle(_ context.Context, r slog.Record) error {
	// Choose color based on level
	var levelColor *color.Color
	switch r.Level {
	case slog.LevelDebug:
		levelColor = color.New(color.FgHiCyan)
	case slog.LevelInfo:
		levelColor = color.New(color.FgHiGreen)
	case slog.LevelWarn:
		levelColor = color.New(color.FgHiYellow)
	case slog.LevelError:
		levelColor = color.New(color.FgHiRed)
	default:
		levelColor = color.New(color.FgWhite)
	}

	// Format timestamp
	timestamp := color.HiBlackString(r.Time.Format("15:04:05"))
	levelStr := levelColor.Sprintf("[%s]", strings.ToUpper(r.Level.String()))
	msgStr := levelColor.Sprintf("%s", r.Message)

	// Collect structured attributes inline
	attrParts := make([]string, 0, r.NumAttrs())
	r.Attrs(func(attr slog.Attr) bool {
		key := color.HiCyanString(attr.Key)
		value := formatValueInline(attr.Value)
		attrParts = append(attrParts, fmt.Sprintf("%s=%s", key, value))
		return true
	})

	attrsStr := ""
	if len(attrParts) > 0 {
		attrsStr = " " + strings.Join(attrParts, " ")
	}

	fmt.Printf("%s %s %s%s\n", timestamp, levelStr, msgStr, attrsStr)
	return nil
}

func formatValueInline(value slog.Value) string {
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

// Setup sets up and installs the logger as the global default slog logger.
func Setup(opt *Options) error {
	level, err := opt.GetLevelValue()
	if err != nil {
		return err
	}

	var handler slog.Handler
	if opt.Production {
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
		baseHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     level,
			AddSource: opt.AddSource,
		})
		handler = &colorfulHandler{
			handler: baseHandler,
			level:   level,
		}
	}

	slog.SetDefault(slog.New(handler))
	return nil
}

func SetupDefault() {
	Setup(&Options{
		AddSource:  false,
		Production: false,
		Level:      "info",
	})
}
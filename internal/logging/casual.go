package logging

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"strings"
	"unicode"

	"github.com/fatih/color"
)

// CasualHandler for CLI mode with colors
type CasualHandler struct {
	slog.Handler
	logger *log.Logger
	level  slog.Level
}

func NewCliHandler(level slog.Level) *CasualHandler {
	return &CasualHandler{
		Handler: slog.NewTextHandler(
			os.Stdout, &slog.HandlerOptions{
				Level:     level,
				AddSource: level == slog.LevelDebug,
			},
		),
		logger: log.New(os.Stdout, "", 0),
		level:  level,
	}
}

func (h *CasualHandler) Handle(ctx context.Context, record slog.Record) error {
	levelString, message := h.formatLevelAndMessage(record)
	formattedFields := h.formatFields(record)

	// Show metadata only if level is DEBUG
	if h.level == slog.LevelDebug {
		h.printWithMetadata(levelString, message, formattedFields, record)
	} else {
		h.printWithoutMetadata(message, formattedFields)
	}

	return nil
}

func (h *CasualHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

func (h *CasualHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CasualHandler{
		Handler: h.Handler.WithAttrs(attrs),
		logger:  h.logger,
		level:   h.level,
	}
}

func (h *CasualHandler) WithGroup(name string) slog.Handler {
	return &CasualHandler{
		Handler: h.Handler.WithGroup(name),
		logger:  h.logger,
		level:   h.level,
	}
}

func (h *CasualHandler) formatLevelAndMessage(record slog.Record) (string, string) {
	levelString := record.Level.String()
	message := record.Message

	// Capitalize first character if it's lowercase
	if len(message) > 0 && unicode.IsLower(rune(message[0])) {
		message = string(unicode.ToUpper(rune(message[0]))) + message[1:]
	}

	// Apply colors based on level
	levelString, message = colorizeByLevel(record.Level, levelString, message)

	return levelString, message
}

func (h *CasualHandler) formatFields(record slog.Record) string {
	fieldCount := record.NumAttrs()
	if fieldCount == 0 {
		return ""
	}

	fields := make(map[string]any, fieldCount)
	record.Attrs(func(attr slog.Attr) bool {
		fields[attr.Key] = attr.Value.Any()
		return true
	})

	if len(fields) == 0 {
		return ""
	}

	// For single field, format as {"key": value}
	if len(fields) == 1 {
		for key, value := range fields {
			valueBytes, err := json.Marshal(value)
			if err != nil {
				return color.RedString("[ERROR marshaling fields]")
			}
			jsonStr := "{\"" + key + "\": " + string(valueBytes) + "}"
			return color.MagentaString(jsonStr)
		}
	}

	// Multiple fields, use indented formatting with newline prefix
	jsonBytes, err := json.MarshalIndent(fields, "", "  ")
	if err != nil {
		return color.RedString("[ERROR marshaling fields]")
	}

	return color.MagentaString("\n" + string(jsonBytes))
}

func (h *CasualHandler) printWithMetadata(levelString, message, formattedFields string, record slog.Record) {
	timeString := record.Time.Format("15:04:05.000")
	sourceInfo := extractSourceInfo(record)

	// Apply color to metadata based on level
	timeString = colorizeString(record.Level, timeString)
	if sourceInfo != "" {
		sourceInfo = colorizeString(record.Level, sourceInfo)
	}

	// Build log line with pipe separators
	var logParts []string
	logParts = append(logParts, levelString, timeString)
	if sourceInfo != "" {
		logParts = append(logParts, sourceInfo)
	}
	logParts = append(logParts, message)

	logLine := strings.Join(logParts, " | ")

	if formattedFields != "" {
		h.logger.Println(logLine, formattedFields)
	} else {
		h.logger.Println(logLine)
	}
}

func (h *CasualHandler) printWithoutMetadata(message, formattedFields string) {
	if formattedFields != "" {
		h.logger.Println(message, formattedFields)
	} else {
		h.logger.Println(message)
	}
}

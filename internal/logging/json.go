package logging

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
)

// JSONHandler for server mode with optional coloring
type JSONHandler struct {
	opts      *slog.HandlerOptions
	writer    io.Writer
	attrs     []slog.Attr
	groups    []string
	useColors bool
}

func NewJSONHandler(level slog.Level, useColors bool) *JSONHandler {
	return &JSONHandler{
		opts: &slog.HandlerOptions{
			Level:     level,
			AddSource: true, // Always add source for JSON
		},
		writer:    os.Stdout,
		attrs:     []slog.Attr{},
		groups:    []string{},
		useColors: useColors,
	}
}

func (h *JSONHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *JSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &JSONHandler{
		opts:      h.opts,
		writer:    h.writer,
		attrs:     newAttrs,
		groups:    h.groups,
		useColors: h.useColors,
	}
}

func (h *JSONHandler) WithGroup(name string) slog.Handler {
	newGroups := make([]string, len(h.groups)+1)
	copy(newGroups, h.groups)
	newGroups[len(h.groups)] = name

	return &JSONHandler{
		opts:      h.opts,
		writer:    h.writer,
		attrs:     h.attrs,
		groups:    newGroups,
		useColors: h.useColors,
	}
}

func (h *JSONHandler) Handle(ctx context.Context, r slog.Record) error {
	fields := h.buildFields(r)

	jsonBytes, err := json.Marshal(fields)
	if err != nil {
		return err
	}

	output := h.formatOutput(jsonBytes, r.Level)

	_, err = h.writer.Write(output)
	return err
}

func (h *JSONHandler) buildFields(r slog.Record) map[string]interface{} {
	fields := make(map[string]interface{})
	fields["level"] = r.Level.String()
	fields["time"] = r.Time.Format("2006-01-02T15:04:05.000Z07:00")

	// Add source if available
	if sourceInfo := extractSourceInfo(r); sourceInfo != "" {
		fields["source"] = sourceInfo
	}

	fields["msg"] = r.Message

	// Add pre-configured attributes
	for _, attr := range h.attrs {
		fields[attr.Key] = attr.Value.Any()
	}

	// Add record attributes
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})

	return fields
}

func (h *JSONHandler) formatOutput(jsonBytes []byte, level slog.Level) []byte {
	if !h.useColors {
		return append(jsonBytes, '\n')
	}

	jsonStr := string(jsonBytes)
	coloredStr := colorizeString(level, jsonStr)
	return append([]byte(coloredStr), '\n')
}

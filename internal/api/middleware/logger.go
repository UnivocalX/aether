package middleware

import (
	"context"
	"strings"

	"log/slog"

	"github.com/gin-gonic/gin"
)

// --- Forward Gin internal logs & route debug into slog so formatting follows your logger ---
// This runs when the package is initialized. It only sets writers/funcs; actual logging
// will occur later (after your logger.Setup runs in cmd.Setup).
func init() {
	gin.DefaultWriter = &slogWriter{}
	gin.DefaultErrorWriter = &slogWriter{}

	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, numHandlers int) {
		attrs := []slog.Attr{
			slog.String("method", httpMethod),
			slog.String("path", absolutePath),
			slog.String("handler", handlerName),
			slog.Int("handlers", numHandlers),
		}
		// route debug is low-level; use Debug level
		slog.LogAttrs(context.Background(), slog.LevelDebug, "route", attrs...)
	}
}

// slogWriter forwards Gin's internal write output into slog so it uses your installed handler.
type slogWriter struct{}

func (w *slogWriter) Write(p []byte) (int, error) {
	line := strings.TrimSpace(string(p))
	if line == "" {
		return len(p), nil
	}

	// Strip common Gin prefixes to keep messages concise
	line = strings.TrimPrefix(line, "[GIN-debug] ")
	line = strings.TrimPrefix(line, "[GIN] ")
	line = strings.TrimSpace(line)

	// Heuristic level selection
	switch {
	case strings.Contains(line, "ERROR") || strings.Contains(line, "[ERROR]") || strings.Contains(line, "panic"):
		slog.Error(line)
	case strings.Contains(line, "WARN") || strings.Contains(line, "[WARNING]"):
		slog.Warn(line)
	default:
		// Gin's normal informational messages
		slog.Info(line)
	}

	return len(p), nil
}

// logging middle ware logs HTTP requests using slog.LogAttrs which respects default handler
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Common attrs (no request_id for now)
		common := []slog.Attr{
			slog.String("client_ip", getClientIP(c)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.RequestURI),
		}

		// Start (Debug so it can be silenced in production)
		slog.LogAttrs(c.Request.Context(), slog.LevelDebug, "HTTP Request Started", common...)

		// Let handlers run
		c.Next()
	}
}

// getClientIP gets the client IP address
func getClientIP(c *gin.Context) string {
	// Try different headers for client IP
	if ip := c.Request.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := c.Request.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return c.Request.RemoteAddr
}

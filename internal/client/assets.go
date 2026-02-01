package client

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"path/filepath"

	"github.com/UnivocalX/aether/pkg/universe"
)

func LoadAssets(pattern string, progress bool) error {
	slog.Info("starting to load files as assets", "pattern", pattern)

	// glob pattern matches
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if matches == nil {
		return fmt.Errorf("failed to find files that match the pattern")
	}
	slog.Info("found candidates", "total", len(matches))

	// analyze directory files
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	files := analyzePipeline(ctx, matches, progress)
	universe.Drain(ctx, files)

	if ctx.Err() != nil {
		defer fmt.Printf("\n")
		return fmt.Errorf("load: %w", ctx.Err())
	}

	slog.Info("successfully loaded assets")
	return nil
}
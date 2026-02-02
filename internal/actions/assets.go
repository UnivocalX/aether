package actions

import (
	"context"
	"fmt"
	"log/slog"

	"path/filepath"
)

func (c *Client) LoadAssets(pattern string) error {
	slog.Info("starting to load files as assets", "pattern", pattern)

	// find matches
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("no files matched the given pattern")
	}
	slog.Info("found candidates", "total", len(matches))

	// run analyze
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	stream := analyzePipeline(ctx, matches, c.interactive)
	success, _, err := handleAnalysisResult(ctx, stream, c.interactive)

	if err != nil {
		return fmt.Errorf("load: %w", err)
	}

	slog.Info("successfully loaded assets", "total", len(success))
	return nil
}


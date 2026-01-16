package pipelines

import (
	"context"
	"fmt"
	"log/slog"

	"path/filepath"

	"github.com/UnivocalX/aether/pkg/universe"
)


func GenerateManifestPipeline(pattern string, manifestPath string) error {
	slog.Info("starting to generate manifest file", "pattern", pattern)

	// Find matches
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("generate: %w %q", err, pattern)
	} 

	if len(matches) == 0 {
		return fmt.Errorf("generate: no matching files %q", pattern)
	}

	context.
	universe.NewPipeline()
}
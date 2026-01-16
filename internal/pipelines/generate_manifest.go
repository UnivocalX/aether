package pipelines

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"crypto/sha256"
	"path/filepath"

	"github.com/UnivocalX/aether/pkg/universe"
)

func checksum(path string) (string, error) {
	slog.Debug("computing file checksum")

	// attempt to read file
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// compute checksum
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("checksum: %w %q", err, path)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func hasValue(value string) bool {
	if value == "" {
		return false
	}

	return true
}

func errorObserver[T any](env universe.Envelope[T]) {
	slog.Error(env.Err.Error())
}

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

	// Create pipeline
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	checksumCalculator := universe.ValueTransformer(checksum)
	successPredicate := universe.ValuePredicate(hasValue)

	pipeline := universe.NewPipeline(
		universe.Concurrent(universe.Map(checksumCalculator), 8),
		universe.Tap[string](errorObserver),
		universe.Filter(successPredicate),
		universe.UntilDone[string](),
	)

	// Run pipeline
	source := universe.Source(ctx, matches...)
	checksums := pipeline.Run(ctx, source)
}

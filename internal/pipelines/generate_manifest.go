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

	checksumCalculator := universe.TransformAdapter(checksum)
	successPredicate := universe.PredicateAdapter(hasValue)
	log := universe.ObserveErrorAdapter[string](
		func(e error) {
			slog.Error(e.Error())
		},
	)
	consumer := universe.ConsumeAdapter(
		func(checksum string) error {
			slog.Info("computed checksum", "sha256", checksum)
			return nil
		},
	)

	pipeline := universe.NewPipeline(
		universe.Concurrent(universe.Map(checksumCalculator), 8),
		universe.Tap(log),
		universe.Filter(successPredicate),
		universe.UntilDone[string](),
	)

	// Run pipeline
	source := universe.Source(ctx, matches...)
	checksumStream := pipeline.Run(ctx, source)

	if err := universe.Consume[string](
		ctx,
		checksumStream,
		consumer,
	); err != nil {
		return err
	}

	return nil
}

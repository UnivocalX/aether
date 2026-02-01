package client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"log/slog"

	"github.com/UnivocalX/aether/pkg/universe"
	"github.com/schollz/progressbar/v3"
)

// FileChecksum represents a file and its SHA256 checksum.
type FileChecksum struct {
	Path     string `yaml:"path"`
	Checksum string `yaml:"checksum" binding:"required,len=64,hexadecimal"`
}

// Analyze a single file and return its FileChecksum.
func analyzeFile(path string) (FileChecksum, error) {
	slog.Debug("analyzing file", "path", path)
	fc := FileChecksum{Path: path}

	// resolve abs path
	abs, err := filepath.Abs(path)
	if err != nil {
		return fc, err
	}
	fc.Path = abs

	// compute checksum
	c, err := checksum(abs)
	if err != nil {
		return fc, err
	}
	fc.Checksum = c

	slog.Debug("analyze completed", "checksum", c, "path", abs)
	return fc, nil
}

// Compute SHA256 checksum of a file.
func checksum(path string) (string, error) {
	slog.Debug("computing file checksum", "path", path)

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("checksum: %w %q", err, path)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func analyzePipeline(ctx context.Context, paths []string, progress bool) <-chan universe.Envelope[FileChecksum] {
	slog.Info("starting to analyze paths...", "total", len(paths))
	fmt.Printf("\n")

	// transform path string to FileChecksum
	analyzer := func(env universe.Envelope[string]) universe.Envelope[FileChecksum] {
		fc, err := analyzeFile(env.Value)
		return universe.Envelope[FileChecksum]{
			Value: fc,
			Err:   err,
		}
	}

	// create progress bar
	bar := progressbar.NewOptions(
		len(paths),
		progressbar.OptionSetDescription("Analyzing"),
		progressbar.OptionThrottle(200*time.Millisecond),
		progressbar.OptionSetVisibility(progress),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func ()  {
			defer fmt.Printf("\n")
		}),
	)

	// handle progress bar progress
	barHandler := func(env universe.Envelope[FileChecksum]) {
		if env.Err != nil {
			slog.Error("analyze failed", "error", env.Err)
		}
		bar.Add(1)
	}

	// build pipeline
	source := universe.From(ctx, paths...)
	pipeline := universe.Map(source, analyzer, 8).Tap(barHandler, 1)
	return pipeline.Run(ctx)
}

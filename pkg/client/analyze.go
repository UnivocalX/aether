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

// fileAnalysis represents a file and its SHA256 checksum.
type fileAnalysis struct {
	Path     string `yaml:"path"`
	Checksum string `yaml:"checksum" binding:"required,len=64,hexadecimal"`
}

// Analyze a single file and return its FileChecksum.
func analyzeFile(path string) (fileAnalysis, error) {
	slog.Debug("analyzing file", "path", path)
	fc := fileAnalysis{Path: path}

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

func analyzePipeline(ctx context.Context, paths []string, progress bool) universe.Stream[fileAnalysis] {
	slog.Info("starting to analyze paths...", "total", len(paths))

	analyzer := universe.TransformValue(analyzeFile)

	// create progress bar
	bar := progressbar.NewOptions(
		len(paths),
		progressbar.OptionSetDescription("Analyzing"),
		progressbar.OptionThrottle(200*time.Millisecond),
		progressbar.OptionSetVisibility(progress),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionOnCompletion(func() {
			defer fmt.Printf("\n")
		}),
	)

	if progress {
		fmt.Fprintln(os.Stdout)
		bar.RenderBlank()
		bar.Clear()
	}

	// handle progress bar progress
	barHandler := func(meta *universe.Meta, env universe.Envelope[fileAnalysis]) {
		if env.Err != nil {
			slog.Error("analyze failed", "error", env.Err)
		}
		bar.Add(1)
	}

	// build pipeline
	source := universe.From(ctx, paths...)
	return universe.Map(source, analyzer, 8).Tap(barHandler, 1).Run(ctx)
}

type AssetsAnalysis struct {
	success <-chan universe.Envelope[fileAnalysis]
	failure universe.Envelope[fileAnalysis]
}

// handleAnalysisResult partitions a stream of file analyses into successes and failures.
// If both exist and interactive is true, prompts the user whether to continue.
// Returns slices of successful and failed analyses, plus an error if partitioning failed
// or the user aborted.
func handleAnalysisResult(
	ctx context.Context,
	stream universe.Stream[fileAnalysis],
	interactive bool,
) (
	[]universe.Envelope[fileAnalysis],
	[]universe.Envelope[fileAnalysis],
	error,
) {
	success, failure, err := universe.Partition(ctx, stream.Data)

	switch {
	// Return immediately if partitioning the stream failed
	case err != nil:
		return nil, nil, err

	// If no file was analyzed successfully, return failures and an error
	case len(success) == 0:
		return success, failure, fmt.Errorf("no files were successfully analyzed")

	// If all files were analyzed successfully, return the success list
	case len(failure) == 0:
		return success, failure, nil

	// If there are both successes and failures, ask the user whether to continue
	default:
		prompt := fmt.Sprintf(
			"failed to analyze %d out of %d files. Continue?",
			len(failure),
			len(success)+len(failure),
		)

		approved, err := Input(ctx, prompt, interactive)
		if err != nil {
			return success, failure, err
		}

		if !approved {
			return success, failure, fmt.Errorf(
				"aborted due to %d analysis failures",
				len(failure),
			)
		}

		return success, failure, nil
	}
}

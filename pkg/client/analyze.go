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

type AnalysisResults struct {
	failures  map[string]universe.Envelope[fileAnalysis]
	successes map[string]universe.Envelope[fileAnalysis]
}

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

func handleAnalysisResult(
	ctx context.Context,
	stream universe.Stream[fileAnalysis],
	interactive bool,
) (AnalysisResults, error) {
	successList, failureList, err := universe.Partition(ctx, stream.Data)

	successes := make(map[string]universe.Envelope[fileAnalysis], len(successList))
	for _, env := range successList {
		successes[env.Value.Checksum] = env
	}

	failures := make(map[string]universe.Envelope[fileAnalysis], len(failureList))
	for _, env := range failureList {
		failures[env.Value.Path] = env
	}

	result := AnalysisResults{
		successes: successes,
		failures:  failures,
	}

	switch {
	case err != nil:
		return result, err

	case len(successes) == 0:
		return result, fmt.Errorf("no files were successfully analyzed")

	case len(failures) == 0:
		return result, nil

	default:
		prompt := fmt.Sprintf(
			"failed to analyze %d out of %d files. Continue?",
			len(failures),
			len(successes)+len(failures),
		)

		approved, err := Input(ctx, prompt, interactive)
		if err != nil {
			return result, err
		}

		if !approved {
			return result, fmt.Errorf(
				"aborted due to %d analysis failures",
				len(failures),
			)
		}
		return result, nil
	}
}

// Summary
// Issue  						| Impact at 10k 			| Fix 
// Sequential uploads 			| High major bottleneck 	| Parallelize like analysis
// All responses in memory 		| Low — manageable 			| Merge post+upload per batch
// No resume/retry 				| Medium 					| Track completed checksums 
// No file size guard 			| Low — edge case 			| os.Stat before hashing
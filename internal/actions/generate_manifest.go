package actions

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
	yaml "gopkg.in/yaml.v3"
)

// Manifest holds a list of file checksums with thread-safe access.
type Manifest struct {
	Files []*FileChecksum `yaml:"files"`
}

// Write manifest to YAML file
func (m *Manifest) Write(path string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("generate: failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("generate: failed to write manifest to %q: %w", path, err)
	}

	return nil
}

// FileChecksum represents a file and its SHA256 checksum.
type FileChecksum struct {
	Path     string `yaml:"path"`
	Checksum string `yaml:"checksum" binding:"required,len=64,hexadecimal"`
}

// Analyze a single file and return its FileChecksum.
func analyzeFile(path string) (*FileChecksum, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	c, err := checksum(abs)
	if err != nil {
		return nil, err
	}

	return &FileChecksum{Path: abs, Checksum: c}, nil
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

func analyzeDirectoryPipeline(ctx context.Context, paths []string) *universe.Pipeline[*FileChecksum] {
	// transform path string to FileChecksum - work directly with Envelope
	analyzer := func(env universe.Envelope[string]) universe.Envelope[*FileChecksum] {
		if env.Err != nil {
			return universe.Envelope[*FileChecksum]{
				Value: nil,
				Err:   env.Err,
			}
		}

		fc, err := analyzeFile(env.Value)
		return universe.Envelope[*FileChecksum]{
			Value: fc,
			Err:   err,
		}
	}

	// filter only valid checksums - work directly with Envelope
	hasValidChecksum := func(env universe.Envelope[*FileChecksum]) bool {
		return env.Err == nil && env.Value != nil && env.Value.Checksum != ""
	}

	// log errors - work directly with Envelope
	logErrors := func(env universe.Envelope[*FileChecksum]) {
		if env.Err != nil {
			slog.Error("failed to analyze file", "error", env.Err.Error())
		}
	}

	// build pipeline
	source := universe.From(ctx, paths...)
	pipeline := universe.Map(source, analyzer, 8).
		Tap(logErrors, 1).
		Filter(hasValidChecksum, 1).
		UntilDone()

	return pipeline
}

func GenerateManifest(pattern string, manifestPath string) error {
	slog.Info("starting to generate manifest file", "pattern", pattern)

	// Find matches
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("generate: %w %q", err, pattern)
	}
	if len(matches) == 0 {
		return fmt.Errorf("generate: no matching files %q", pattern)
	}

	// Run pipeline
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	checksums := analyzeDirectoryPipeline(ctx, matches).Run(ctx)

	// generate manifest - work directly with Envelope
	manifestReducer := func(m *Manifest, env universe.Envelope[*FileChecksum]) *Manifest {
		if env.Err == nil && env.Value != nil {
			m.Files = append(m.Files, env.Value)
		}
		return m
	}

	manifest, err := universe.Reduce(
		ctx,
		checksums,
		manifestReducer,
		&Manifest{},
	)

	if err != nil {
		return fmt.Errorf("generate: failed to build manifest: %w", err)
	}

	// write manifest to file
	if err := manifest.Write(manifestPath); err != nil {
		return err
	}

	slog.Info("manifest generated successfully", "path", manifestPath, "files", len(manifest.Files))
	return nil
}

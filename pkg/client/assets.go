package client

import (
	"context"
	"fmt"
	"log/slog"

	"path/filepath"

	v1 "github.com/UnivocalX/aether/pkg/web/api/handlers/v1"
	"github.com/UnivocalX/aether/pkg/universe"
)

const (
	BatchSize = 1000
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
		return err
	}

	// post assets in batches
	submitAssets(success)

	slog.Info("successfully loaded assets", "total", len(success))
	return nil
}

func submitAssets(analysis []universe.Envelope[fileAnalysis]) {
	slog.Info("submitting new assets")

	// submit in batches
	for start := 0; start < len(analysis); start += BatchSize {
		// calculate batch size
		end := start + BatchSize
		if end > len(analysis) {
			end = len(analysis)
		}

		batchEnvelopes := analysis[start:end]
		files := make([]fileAnalysis, len(batchEnvelopes))
		for i, env := range batchEnvelopes {
			files[i] = env.Value
		}

		request, err := newAssetsBatchRequest(files)
	}
}

// create new assets batch request
func newAssetsBatchRequest(files []fileAnalysis) (*v1.AssetsBatchResponse, error) {
	var request v1.CreateAssetsBatchRequest
	assets := make([]v1.AssetPayload, len(files))

	for i, f := range files {
		asset := v1.AssetPayload{
			Checksum: f.Checksum,
			Display:  filepath.Base(f.Path),
		}
		assets[i] = asset
	}
	request.Assets = assets

	return nil, nil
}

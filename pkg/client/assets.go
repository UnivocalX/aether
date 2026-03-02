package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"path/filepath"

	"github.com/UnivocalX/aether/pkg/universe"
	"github.com/UnivocalX/aether/pkg/web/api/dto"
	v1 "github.com/UnivocalX/aether/pkg/web/api/handlers/v1"
)

const (
	BatchSize = 1000
)

func (c *Client) LoadAssets(ctx context.Context, pattern string) error {
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

	// analyze candidates
	stream := analyzePipeline(ctx, matches, c.interactive)
	success, _, err := handleAnalysisResult(ctx, stream, c.interactive)
	if err != nil {
		return err
	}

	// post assets in batches
	analysis := envelopesToAnalysis(success...)
	c.postAssets(ctx, analysis)

	slog.Info("successfully loaded assets", "total", len(success))
	return nil
}

func (c *Client) postAssets(ctx context.Context, analysis []fileAnalysis) error {
	slog.Info("posting new assets")

	// submit in batches
	for start := 0; start < len(analysis); start += BatchSize {
		// calculate batch size
		end := start + BatchSize
		if end > len(analysis) {
			end = len(analysis)
		}

		batch := analysis[start:end]
		response, err := c.postAssetsBatch(ctx, batch)
		if err != nil {
			return err
		}
	}
}

func (c *Client) postAssetsBatch(ctx context.Context, analysis []fileAnalysis) (*v1.AssetsBatchResponse, error) {
	assets := make([]v1.AssetPayload, len(analysis))
	for i, f := range analysis {
		assets[i] = v1.AssetPayload{
			Checksum: f.Checksum,
			Display:  filepath.Base(f.Path),
		}
	}

	b, err := json.Marshal(v1.CreateAssetsBatchRequest{Assets: assets})
	if err != nil {
		return nil, err
	}

	resp, err := c.post(ctx, "/batch/assets", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, decodeErrorResponse(resp)
	}

	var response v1.AssetsBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func envelopesToAnalysis(envelopes ...universe.Envelope[fileAnalysis]) []fileAnalysis {
	slog.Debug("extracting analysis from envelopes")
	var analysis []fileAnalysis
	for _, env := range envelopes {
		analysis = append(analysis, env.Value)
	}
	return analysis
}

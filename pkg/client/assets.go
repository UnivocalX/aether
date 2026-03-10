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
	v1 "github.com/UnivocalX/aether/pkg/web/api/handlers/v1"
)

const (
	BatchSize          = 1000
	AssetsBatchApiPath = "/batch/assets"
)

// matches should get a list of file path so move glob outside
// load should return success, failure and error for easier retry outside of it, if needed
// stream the analysis results as a input to a universe pipeline that has 2 stages, 1. Post 2. Upload
func (c *Client) LoadAssets(ctx context.Context, pattern string) error {
	slog.Info("starting to load files as assets", "pattern", pattern)

	// resolve matches
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("no files matched the given pattern")
	}
	slog.Info("found candidates", "total", len(matches))

	// analyze matches (validate file, get checksum, resolve full path)
	stream := analyzePipeline(ctx, matches, c.durable)

	// fail if not durable
	// if durable, continue
	// return success and failure

	success, _, err := handleAnalysisResult(ctx, stream, c.durable)
	if err != nil {
		return err
	}

	// post assets
	responses, err := c.PostAssets(ctx, assets...)
	if err != nil {
		return err
	}

	// upload assets
	if err := c.UploadAssets(ctx, result, responses...); err != nil {
		return err
	}

	slog.Info("successfully loaded assets", "total", len(result.successes))
	return nil
}

// PostAssets splits assets into batches and posts each one.
func (c *Client) postAssets(ctx context.Context, analysis ...universe.Envelope[fileAnalysis]) ([]*v1.AssetsBatchResponse, error) {
	slog.Info("posting assets", "total", len(assets))

	responses := make([]*v1.AssetsBatchResponse, 0, len(assets)/BatchSize+1)
	for start := 0; start < len(assets); start += BatchSize {
		end := min(start+BatchSize, len(assets))

		batch := v1.CreateAssetsBatchRequest{Assets: assets[start:end]}
		batchResp, err := c.PostAssetsBatch(ctx, batch)
		if err != nil {
			return nil, err
		}
		responses = append(responses, batchResp)
	}

	return responses, nil
}

func (c *Client) PostAssetsBatch(ctx context.Context, req v1.CreateAssetsBatchRequest) (*v1.AssetsBatchResponse, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.post(ctx, AssetsBatchApiPath, bytes.NewReader(b))
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

// UploadAssets uploads each file to its corresponding ingress URL, matched by checksum.
func (c *Client) UploadAssets(ctx context.Context, result AnalysisResults, responses ...*v1.AssetsBatchResponse) error {
	for _, response := range responses {
		for _, asset := range response.Assets {
			env, ok := result.successes[asset.Checksum]
			if !ok {
				return fmt.Errorf("no local file found for checksum %s", asset.Checksum)
			}
			if _, err := c.upload2Storage(ctx, env.Value.Path, asset.IngressUrl); err != nil {
				return fmt.Errorf("failed to upload asset %s: %w", env.Value.Path, err)
			}
		}
	}
	return nil
}

// Summary
// Issue  						| Impact at 10k 			| Fix
// Sequential uploads 			| High major bottleneck 	| Parallelize like analysis
// All responses in memory 		| Low — manageable 			| Merge post+upload per batch
// No resume/retry 				| Medium 					| Track completed checksums
// No file size guard 			| Low — edge case 			| os.Stat before hashing

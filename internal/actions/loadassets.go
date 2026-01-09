package actions

import (
	// "encoding/json"
	"fmt"
	// // "io"
	// "log/slog"
	// "os"
	// "path/filepath"

	// "github.com/UnivocalX/aether/internal/helpers"
	// v1 "github.com/UnivocalX/aether/internal/web/api/handlers/v1"
)

type LoadAssets struct {
	Connector
}

func NewLoadAssets(opts ...ConnectorOption) (*LoadAssets, error) {
	connector, err := NewConnector(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create connector: %w", err)
	}

	return &LoadAssets{
		Connector: *connector,
	}, nil
}

func (load *LoadAssets) Run(patterns []string, tags []string) error {
	// slog.Info("running load assets")

	// // collect all matching files from patterns
	// matchedFiles := load.collectMatchingFiles(patterns)
	// if len(matchedFiles) == 0 {
	// 	return fmt.Errorf("found 0 matching files")
	// }

	// slog.Info("successfully collected files", "total", len(matchedFiles))

	// // process files
	// createRequests := load.processFiles(matchedFiles)
	// if len(createRequests) <= 0 {
	// 	return fmt.Errorf("failed to process potential assets")
	// }

	// Resolve tag names to IDs
	return nil
}

// // func (load *LoadAssets) processFiles(files []string) []*v1.AssetPostRequest {
// // 	slog.Info("attempting to process files")

// // 	act := NewAction[string, string]()
// // 	artifacts := act.Run(load.processFile, files...)

// // 	var requests []*v1.AssetPostRequest
// // 	for a := range artifacts {
// // 		sha256, path, err := a.Unwrap()

// // 		if err != nil {
// // 			slog.Error("failed to process file", "path", path)
// // 			continue
// // 		}

// // 		slog.Debug("adding new request", "path", path, "sha256", sha256)
// // 		requests = append(
// // 			requests, &v1.AssetPostRequest{
// // 				AssetUriParams: v1.AssetUriParams{
// // 					Checksum: sha256,
// // 				},
// // 				AssetPostPayload: v1.AssetPostPayload{
// // 					Display: filepath.Base(path),
// // 				},
// // 			},
// // 		)
// // 	}

// // 	slog.Debug("finished processing files", "total", len(requests))
// // 	return requests
// // }

// func (load *LoadAssets) processFile(path string) (string, error) {
// 	slog.Debug("processing potential asset", "path", path)
// 	return helpers.CalculateSHA256(path)
// }

// func (load *LoadAssets) collectMatchingFiles(patterns []string) []string {
// 	slog.Info("collecting files...")
// 	slog.Debug("attempting to expend patterns", "totalPatterns", len(patterns))

// 	var allFiles []string
// 	for _, pattern := range patterns {

// 		// Resolve each pattern to concrete file paths
// 		action := NewAction[string, []string]()
// 		artifacts := action.Run(load.expandPattern, pattern)

// 		for a := range artifacts {
// 			files, _, err := a.Unwrap()

// 			if err != nil {
// 				slog.Error("pattern resolution failed", "pattern", pattern, "error", err)
// 				continue
// 			}

// 			slog.Debug("pattern resolved", "pattern", pattern, "filesFound", len(files))
// 			allFiles = append(allFiles, files...)
// 		}
// 	}

// 	slog.Debug("file collection completed", "totalFiles", len(allFiles))
// 	return allFiles
// }

// func (load *LoadAssets) expandPattern(pattern string) ([]string, error) {
// 	slog.Debug("expanding pattern to files", "pattern", pattern)

// 	// Check if pattern refers to file or directory
// 	fileInfo, err := os.Stat(pattern)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to stat path: %w", err)
// 	}

// 	var files []string
// 	if fileInfo.IsDir() {
// 		// Collect all regular files from directory
// 		entries, err := os.ReadDir(pattern)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to read directory: %w", err)
// 		}

// 		for _, entry := range entries {
// 			if entry.Type().IsRegular() {
// 				files = append(files, filepath.Join(pattern, entry.Name()))
// 			}
// 		}

// 		slog.Debug("directory expanded", "directory", pattern, "filesFound", len(files))
// 	} else {
// 		// Single file
// 		files = append(files, pattern)
// 		slog.Debug("file pattern resolved", "file", pattern)
// 	}

// 	return files, nil
// }

// // func (act *LoadAssets) resolveTags(names []string) error {
// // 	act.tags = make([]uint, len(names))
// // 	slog.Info("resolving tags", "Total", len(names))

// // 	for i, name := range names {
// // 		tagID, err := act.getTagID(name)
// // 		if err != nil {
// // 			return fmt.Errorf("tag %q: %w", name, err)
// // 		}
// // 		act.tags[i] = tagID
// // 	}

// // 	slog.Debug("resolved tags", "count", len(act.tags))
// // 	return nil
// // }

// // func (act *LoadAssets) getTagID(name string) (uint, error) {
// // 	path := fmt.Sprintf("/%s/tags/%s", act.apiVersion, name)

// // 	// Call Get directly thanks to embedding!
// // 	resp, err := act.Get(path)
// // 	if err != nil {
// // 		return 0, err
// // 	}
// // 	defer resp.Body.Close()

// // 	if resp.StatusCode == http.StatusNotFound {
// // 		return 0, fmt.Errorf("not found")
// // 	}
// // 	if resp.StatusCode != http.StatusOK {
// // 		return 0, fmt.Errorf("status %d", resp.StatusCode)
// // 	}

// // 	body, err := io.ReadAll(resp.Body)
// // 	if err != nil {
// // 		return 0, err
// // 	}

// // 	var tagResp v1.GetTagResponse
// // 	if err := json.Unmarshal(body, &tagResp); err != nil {
// // 		return 0, err
// // 	}

// // 	if tagResp.Tag == nil {
// // 		return 0, fmt.Errorf("empty response")
// // 	}

// // 	slog.Info("tag resolved", "Name", name, "ID", tagResp.Tag.ID)
// // 	return tagResp.Tag.ID, nil
// // }

// // func (act *LoadAssets) Requests() <-chan v1.CreateAssetRequest {
// // 	return act.requests
// // }

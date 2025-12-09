package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/UnivocalX/aether/internal/api/v1/schemas"
	"github.com/UnivocalX/aether/internal/utils"
)

type LoadAssets struct {
	HTTPAction
	requests chan schemas.CreateAssetRequest
	errors   chan error
	apiVersion string
	tags       []uint
}

func NewLoadAssets(ctx context.Context, endpoint string) *LoadAssets {
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	return &LoadAssets{
		HTTPAction: *NewHTTPAction(ctx, endpoint),
		apiVersion: "v1",
	}
}

func (act *LoadAssets) Run(source string, tags ...string) error {
	slog.Debug("executing load assets", "source", source, "tags", len(tags))

	// validate Action
	if err := act.Validate(); err != nil {
		return err
	}

	// gather Assets
	filepaths, err := act.findFiles(source)
	if err != nil {
		return err
	}

	// resolve tags names to ids
	if len(tags) > 0 {
		if err := act.resolveTags(tags); err != nil {
			return err
		}
	}

	// process files and populate requests channel
	if err := act.processFiles(filepaths); err != nil {
		return err
	}

	return nil
}


func (act *LoadAssets) reportErrors() {
    // Drain the channel and count actual errors
    totalErrors := 0
    for err := range act.errors {
        if err != nil {
            slog.Error(err.Error())
            totalErrors++
        }
    }

    // Log the accurate total after draining
    slog.Info("summary report",
        slog.Int("TotalErrors", totalErrors),
        slog.Int("TotalRequests", len(act.requests)),
    )
}


// findFiles returns all regular files matching the pattern
func (act *LoadAssets) findFiles(pattern string) ([]string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("failed to find matches %q", pattern)
	}

	slog.Info("found pattern matches", "Count", len(matches))

	// Use generic parallel processor
	numWorkers := act.CalculateNumOfRoutines(len(matches))
	files, statErrors := utils.ParallelProcess(matches, numWorkers, act.validateFile)

	// Log any stat errors (non-critical)
	if len(statErrors) > 0 {
		for _, err := range statErrors {
			slog.Debug(err.Error())
		}

		return nil, fmt.Errorf("%d files failed stat check, example: %s", len(statErrors), statErrors[0])
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no regular files found in matches %q", pattern)
	}

	slog.Info("found suitable files", "Count", len(files))
	return files, nil
}

func (act *LoadAssets) validateFile(match string) (string, error) {
		info, err := os.Stat(match)
		if err != nil {
			slog.Debug("stat failed", "file", match, "error", err)
			return "", err
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("not a regular file")
		}
		return match, nil
}

// processFiles concurrently processes all files
func (act *LoadAssets) processFiles(filepaths []string) error {
	var wg sync.WaitGroup
	act.requests = make(chan schemas.CreateAssetRequest, len(filepaths))
	act.errors = make(chan error, len(filepaths))

	// convert potential assets list to a queue
	workCh := make(chan string, len(filepaths))
	for _, path := range filepaths {
		workCh <- path
	}
	close(workCh)

	// start parallel processing while splitting the work using a queue
	numWorkers := act.CalculateNumOfRoutines(len(filepaths))
	wg.Add(numWorkers)
	slog.Debug("creating workers", "TotalWorkers", numWorkers)
	slog.Info("starting to process assets")
	for i := 0; i < numWorkers; i++ {
		go act.processFile(&wg, workCh)
	}
	wg.Wait()

	close(act.requests)
	close(act.errors)

	slog.Debug("finished processing assets", "TotalRequests", len(act.requests))

	if len(act.errors) > 0 {
		act.reportErrors()
		return fmt.Errorf("failed to process all suitable assets")
	}
	return nil
}

// processFile processes files from the work channel and sends results to channels
func (act *LoadAssets) processFile(wg *sync.WaitGroup, workCh <-chan string) {
	defer wg.Done()

	slog.Debug("starting work")
	for filePath := range workCh {
		hash, err := utils.CalculateSHA256(filePath)
		if err != nil {
			slog.Debug("failed to calculate sha256", "Filepath", filePath, "Error", err)
			act.errors <- err
			continue
		}

		slog.Debug("updating requests with a new asset", "SHA256", hash, "FilePath", filePath)
		act.requests <- schemas.CreateAssetRequest{
			Display: filePath,
			SHA256:  hash,
			Tags:    act.tags,
		}
	}
	slog.Debug("done processing work")
}

// resolveTags converts tag names to IDs
func (act *LoadAssets) resolveTags(names []string) error {
	u, err := url.Parse(string(act.endpoint))
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, act.apiVersion, "tags")
	baseURL := u.String()

	act.tags = make([]uint, len(names))
	slog.Info("resolving tags", "Total", len(names))
	for i, name := range names {
		act.tags[i], err = act.getTagID(baseURL, name)
		if err != nil {
			return fmt.Errorf("tag %q: %w", name, err)
		}
	}

	slog.Debug("resolved tags", "count", len(act.tags))
	return nil
}

// getTagID fetches a tag ID by name
func (act *LoadAssets) getTagID(baseURL, name string) (uint, error) {
	req, err := http.NewRequestWithContext(act.ctx, "GET", path.Join(baseURL, name), nil)
	if err != nil {
		return 0, err
	}

	resp, err := act.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return 0, fmt.Errorf("not found")
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status %d", resp.StatusCode)
	}

	var tagResp schemas.GetTagResponse
	if err := json.Unmarshal(body, &tagResp); err != nil {
		return 0, err
	}

	if tagResp.Tag == nil {
		return 0, fmt.Errorf("empty response")
	}

	slog.Info("tag resolved", "Name", name, "ID", tagResp.Tag.ID)
	return tagResp.Tag.ID, nil
}

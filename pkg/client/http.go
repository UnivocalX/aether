package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/UnivocalX/aether/pkg/web/api/dto"
)

const (
	maxRetries = 3
	delay      = 1 * time.Second
)

func (c *Client) post(ctx context.Context, path string, body io.ReadSeeker) (*http.Response, error) {
	return c.send(ctx, "POST", path, body)
}

func (c *Client) get(ctx context.Context, path string, body io.ReadSeeker) (*http.Response, error) {
	return c.send(ctx, "GET", path, body)
}

func (c *Client) patch(ctx context.Context, path string, body io.ReadSeeker) (*http.Response, error) {
	return c.send(ctx, "PATCH", path, body)
}

func (c *Client) delete(ctx context.Context, path string, body io.ReadSeeker) (*http.Response, error) {
	return c.send(ctx, "DELETE", path, body)
}

func (c *Client) send(ctx context.Context, method string, path string, body io.ReadSeeker) (*http.Response, error) {
	url := *c.url
	url.Path = path

	// respect timeout per http request
	req, err := http.NewRequestWithContext(ctx, method, url.String(), body)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 && body != nil {
			if _, err = body.Seek(0, io.SeekStart); err != nil {
				return nil, err
			}
		}

		resp, err = c.http.Do(req)
		if !shouldRetry(err, resp) {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}

		// respect timeout between request retries
		select {
		case <-time.After(backoff(attempt)):
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}

	return resp, err
}

func backoff(attempt int) time.Duration {
	base := time.Duration(math.Pow(2, float64(attempt))) * delay
	jitter := time.Duration(rand.Int63n(int64(base / 2)))
	return base + jitter
}

func shouldRetry(err error, resp *http.Response) bool {
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) {
			return netErr.Timeout()
		}
		return true
	}

	code := resp.StatusCode
	return code == http.StatusBadGateway ||
		code == http.StatusServiceUnavailable ||
		code == http.StatusTooManyRequests ||
		code == http.StatusGatewayTimeout
}

func decodeErrorResponse(resp *http.Response) error {
	var errResp dto.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	if errResp.Err != nil {
		return fmt.Errorf("%s: %s", errResp.Msg, errResp.Err.Msg)
	}
	return fmt.Errorf(errResp.Msg)
}
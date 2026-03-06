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
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/UnivocalX/aether/internal/registry"
	"github.com/UnivocalX/aether/pkg/web/api/dto"
)

const (
	maxRetries = 3
	delay      = 1 * time.Second
)

// The caller is responsible for closing the response body.
func (c *Client) get(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(ctx, http.MethodGet, path, body)
	if err != nil {
		return nil, err
	}
	return c.send(req)
}

// The caller is responsible for closing the response body.
func (c *Client) post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	return c.send(req)
}

// The caller is responsible for closing the response body.
func (c *Client) put(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}
	return c.send(req)
}

// The caller is responsible for closing the response body.
func (c *Client) patch(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(ctx, http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}
	return c.send(req)
}

// The caller is responsible for closing the response body.
func (c *Client) delete(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(ctx, http.MethodDelete, path, body)
	if err != nil {
		return nil, err
	}
	return c.send(req)
}

// upload2Storage uploads a file to a presigned URL.
// The caller is responsible for closing the response body.
func (c *Client) upload2Storage(ctx context.Context, presignedUrl registry.Secret, path string) (*http.Response, error) {
	u, err := url.Parse(presignedUrl.Value())
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close() // FIX 1: close the file after the request is done

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u.String(), file)
	if err != nil {
		return nil, err
	}

	req.ContentLength = info.Size()
	req.GetBody = func() (io.ReadCloser, error) {
		_, err := file.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}
		return io.NopCloser(file), nil
	}

	return c.send(req)
}

func (c *Client) send(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// rewind body if possible
		if attempt > 0 && req.GetBody != nil {
			req.Body, err = req.GetBody()
			if err != nil {
				return nil, err
			}
		}

		resp, err = c.http.Do(req)
		if !shouldRetry(err, resp) {
			// success
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		// FIX 3: respect Retry-After header for 429 responses
		waitDur := backoff(attempt)
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
				if secs, parseErr := strconv.Atoi(retryAfter); parseErr == nil {
					waitDur = time.Duration(secs) * time.Second
				}
			}
		}

		select {
		case <-time.After(waitDur):
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}

	return resp, err
}

// newRequest builds an HTTP request targeting the client's base URL at the given path.
func (c *Client) newRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	u := *c.url
	u.Path = path

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, err
	}

	// If body is a file, set Content-Length and GetBody for retries
	if f, ok := body.(*os.File); ok {
		info, statErr := f.Stat()
		if statErr != nil {
			return nil, fmt.Errorf("failed to stat request body file: %w", statErr)
		}
		req.ContentLength = info.Size()
		req.GetBody = func() (io.ReadCloser, error) {
			_, err := f.Seek(0, io.SeekStart)
			if err != nil {
				return nil, err
			}
			return io.NopCloser(f), nil
		}
	}

	// If body supports Seek but is not a file, set GetBody for retries
	if seeker, ok := body.(io.Seeker); ok && req.GetBody == nil {
		req.GetBody = func() (io.ReadCloser, error) {
			_, err := seeker.Seek(0, io.SeekStart)
			if err != nil {
				return nil, err
			}
			return io.NopCloser(body), nil
		}
	}

	return req, nil
}

func backoff(attempt int) time.Duration {
	base := time.Duration(math.Pow(2, float64(attempt))) * delay
	jitter := time.Duration(rand.Int63n(int64(base / 2)))
	return base + jitter
}

func shouldRetry(err error, resp *http.Response) bool {
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false
		}
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
	return errors.New(errResp.Msg)
}
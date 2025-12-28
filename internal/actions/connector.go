package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/UnivocalX/aether/pkg/registry"
)

const (
	defaultHTTPTimeout   = 30 * time.Second
	defaultEndpoint      = "localhost:8080"
	defaultRetryAttempts = 3
	defaultRetryBackoff  = 500 * time.Millisecond
)

// Connector provides HTTP capabilities with retry and timeout support
type Connector struct {
	endpoint      registry.Endpoint
	client        http.Client
	retryAttempts int
	retryBackoff  time.Duration
	ctx           context.Context
}

// ConnectorOption defines a function type that modifies Connector
type ConnectorOption func(*Connector) error

// WithContext sets the context for the connector
func WithContext(ctx context.Context) ConnectorOption {
	return func(c *Connector) error {
		if ctx == nil {
			return fmt.Errorf("context cannot be nil")
		}
		c.ctx = ctx
		slog.Debug("context applied")
		return nil
	}
}

// WithTimeout sets the HTTP client timeout
func WithTimeout(d time.Duration) ConnectorOption {
	return func(c *Connector) error {
		if d <= 0 {
			return fmt.Errorf("timeout must be positive, got: %v", d)
		}
		c.client.Timeout = d
		slog.Debug("http client timeout applied", "timeout", d)
		return nil
	}
}

// WithEndpoint sets the endpoint
func WithEndpoint(ep string) ConnectorOption {
	return func(c *Connector) error {
		ep = strings.TrimSpace(ep)
		if ep == "" {
			return fmt.Errorf("endpoint cannot be empty")
		}
		c.endpoint = registry.Endpoint(ep)
		slog.Debug("endpoint applied", "endpoint", ep)
		return nil
	}
}

// WithRetryAttempts sets the number of retry attempts
func WithRetryAttempts(attempts int) ConnectorOption {
	return func(c *Connector) error {
		if attempts < 0 {
			return fmt.Errorf("retry attempts must be non-negative, got: %d", attempts)
		}
		c.retryAttempts = attempts
		slog.Debug("retry attempts applied", "attempts", attempts)
		return nil
	}
}

// WithRetryBackoff sets the retry backoff duration
func WithRetryBackoff(backoff time.Duration) ConnectorOption {
	return func(c *Connector) error {
		if backoff < 0 {
			return fmt.Errorf("backoff must be non-negative, got: %v", backoff)
		}
		c.retryBackoff = backoff
		slog.Debug("backoff applied", "backoff", backoff)
		return nil
	}
}

// WithHttpClient sets a custom HTTP client
func WithHttpClient(client http.Client) ConnectorOption {
	return func(c *Connector) error {
		c.client = client
		slog.Debug("custom http client applied")
		return nil
	}
}

// NewConnector creates a new Connector with options
func NewConnector(opts ...ConnectorOption) (*Connector, error) {
	// Create connector with defaults
	conn := &Connector{
		endpoint:      registry.Endpoint(defaultEndpoint),
		client:        http.Client{Timeout: defaultHTTPTimeout},
		retryAttempts: defaultRetryAttempts,
		retryBackoff:  defaultRetryBackoff,
		ctx:           context.Background(),
	}

	// Apply all options
	for _, opt := range opts {
		if err := opt(conn); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return conn, nil
}

// HTTP URL building
func (con *Connector) buildURL(path string) (string, error) {
	base := strings.TrimSpace(string(con.endpoint))
	if base == "" {
		base = string(defaultEndpoint)
	}

	if !strings.Contains(base, "://") {
		base = "http://" + base
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint: %w", err)
	}

	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + path

	return baseURL.String(), nil
}

// Error handling utilities
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var ue *url.Error
	if errors.As(err, &ue) {
		err = ue.Err
	}

	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		switch {
		case errors.Is(opErr.Err, syscall.ECONNRESET),
			errors.Is(opErr.Err, syscall.ECONNREFUSED),
			errors.Is(opErr.Err, syscall.EPIPE),
			errors.Is(opErr.Err, syscall.ENETUNREACH),
			errors.Is(opErr.Err, syscall.ECONNABORTED):
			return true
		}
	}

	return false
}

func (con *Connector) checkContextCancelled() error {
	select {
	case <-con.ctx.Done():
		return con.ctx.Err()
	default:
		return nil
	}
}

func (con *Connector) waitWithContext(d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-con.ctx.Done():
		return con.ctx.Err()
	}
}

func shouldRetry(resp *http.Response, err error, attemptNum, maxAttempts int) bool {
	if attemptNum >= maxAttempts {
		return false
	}

	if err != nil {
		return isTransientError(err)
	}

	return resp != nil && resp.StatusCode >= 500
}

func (con *Connector) DoRequestWithRetry(req *http.Request) (*http.Response, error) {
	if req.Context() == nil {
		req = req.WithContext(con.ctx)
	}

	var lastErr error
	backoff := con.retryBackoff

	for attempt := 1; attempt <= con.retryAttempts; attempt++ {
		if err := con.checkContextCancelled(); err != nil {
			slog.Debug("request cancelled before attempt", "attempt", attempt)
			return nil, err
		}

		slog.Debug("attempting request", "attempt", attempt, "method", req.Method, "url", req.URL.String())

		resp, err := con.client.Do(req)

		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}

		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			resp.Body.Close()
		}

		if !shouldRetry(resp, err, attempt, con.retryAttempts) {
			slog.Debug("not retrying", "attempt", attempt, "error", lastErr)
			return nil, lastErr
		}

		slog.Debug("retrying after error", "attempt", attempt, "error", lastErr, "backoff", backoff)

		if err := con.waitWithContext(backoff); err != nil {
			slog.Debug("request cancelled during backoff", "attempt", attempt)
			return nil, err
		}

		backoff *= 2
	}

	slog.Debug("request failed after all attempts", "attempts", con.retryAttempts, "error", lastErr)
	return nil, lastErr
}

// Body preparation
func prepareBody(body any, contentType string) (io.Reader, string, error) {
	if body == nil {
		return nil, contentType, nil
	}

	switch v := body.(type) {
	case io.Reader:
		return v, contentType, nil
	case string:
		return strings.NewReader(v), contentType, nil
	case []byte:
		return bytes.NewReader(v), contentType, nil
	default:
		if contentType == "" {
			contentType = "application/json"
		}
		b, err := json.Marshal(v)
		if err != nil {
			return nil, "", fmt.Errorf("failed to marshal body: %w", err)
		}
		return bytes.NewReader(b), contentType, nil
	}
}

// HTTP methods
func (con *Connector) Get(path string) (*http.Response, error) {
	fullURL, err := con.buildURL(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(con.ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return con.DoRequestWithRetry(req)
}

func (con *Connector) Post(path string, body any, contentType string) (*http.Response, error) {
	fullURL, err := con.buildURL(path)
	if err != nil {
		return nil, err
	}

	bodyReader, finalContentType, err := prepareBody(body, contentType)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(con.ctx, http.MethodPost, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if finalContentType != "" {
		req.Header.Set("Content-Type", finalContentType)
	}

	return con.DoRequestWithRetry(req)
}

func (con *Connector) Put(path string, body any, contentType string) (*http.Response, error) {
	fullURL, err := con.buildURL(path)
	if err != nil {
		return nil, err
	}

	bodyReader, finalContentType, err := prepareBody(body, contentType)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(con.ctx, http.MethodPut, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if finalContentType != "" {
		req.Header.Set("Content-Type", finalContentType)
	}

	return con.DoRequestWithRetry(req)
}

func (con *Connector) Patch(path string, body any, contentType string) (*http.Response, error) {
	fullURL, err := con.buildURL(path)
	if err != nil {
		return nil, err
	}

	bodyReader, finalContentType, err := prepareBody(body, contentType)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(con.ctx, http.MethodPatch, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if finalContentType != "" {
		req.Header.Set("Content-Type", finalContentType)
	}

	return con.DoRequestWithRetry(req)
}

func (con *Connector) Delete(path string) (*http.Response, error) {
	fullURL, err := con.buildURL(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(con.ctx, http.MethodDelete, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return con.DoRequestWithRetry(req)
}

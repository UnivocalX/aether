package client

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	DefaultScheme = "http"
	DefaultHost   = "localhost:8080"
	DefaultPath   = "/health"
	MinTimeout    = 1 * time.Second
	MaxTimeout    = 24 * time.Hour
)

type Client struct {
	durable bool
	url     *url.URL
	http    *http.Client
}

// New creates a new client with options applied and validated
func New(opts ...Option) (*Client, error) {
	c := &Client{
		durable: false,
		url: &url.URL{
			Scheme: DefaultScheme,
			Host:   DefaultHost,
			Path:   DefaultPath,
		},
		http: &http.Client{
			Transport: &http.Transport{

				// Connection Pool
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 200,
				MaxConnsPerHost:     300,

				// Timeouts
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   5 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,

				// Dialer (important under load)
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,

				// Avoid connection churn
				ForceAttemptHTTP2: true,
			},
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// parse and validate url
	if _, e := url.Parse(c.url.String()); e != nil {
		return nil, e
	}

	return c, nil
}

type Option func(*Client) error

// WithHost sets the API host and validates it
func WithHost(h string) Option {
	return func(c *Client) error {
		if h == "" {
			return errors.New("host cannot be empty")
		}
		c.url.Host = h
		return nil
	}
}

// WithHost sets the API host and validates it
func WithScheme(s string) Option {
	return func(c *Client) error {
		if s == "" {
			return errors.New("scheme cannot be empty")
		}
		c.url.Scheme = s
		return nil
	}
}

// WithDurable sets interactive mode
func WithDurable(durable bool) Option {
	return func(c *Client) error {
		c.durable = durable
		return nil
	}
}

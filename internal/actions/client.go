package actions

import (
	"fmt"
	"time"
)

const (
	DefaultEndpoint = "localhost:8080"
	DefaultTimeout  = 30 * time.Minute
)

type Client struct {
	endpoint    string
	timeout     time.Duration
	interactive bool
}

type Option func(*Client) error

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) error {
		c.timeout = timeout
		return nil
	}
}

func WithEndpoint(endpoint string) Option {
	return func(c *Client) error {
		c.endpoint = endpoint
		return nil
	}
}

func Interactive() Option {
	return func(c *Client) error {
		c.interactive = true
		return nil
	}
}

func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		endpoint:    DefaultEndpoint,
		timeout:     DefaultTimeout,
		interactive: false,
	}

	// Apply all options
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}
	return c, nil
}

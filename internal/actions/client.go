package actions

import (
	"errors"
	"fmt"
	"net/url"
	"time"
)

const (
	DefaultEndpoint = "http://localhost:8080"
	DefaultTimeout  = 30 * time.Minute
	MinTimeout      = 1 * time.Second
	MaxTimeout      = 24 * time.Hour
)

type Client struct {
	endpoint    string
	timeout     time.Duration
	interactive bool
}

type Option func(*Client) error

// WithTimeout sets a timeout and validates it
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) error {
		if timeout < MinTimeout || timeout > MaxTimeout {
			return fmt.Errorf("timeout must be between %v and %v", MinTimeout, MaxTimeout)
		}
		c.timeout = timeout
		return nil
	}
}

// WithEndpoint sets the API endpoint and validates it
func WithEndpoint(endpoint string) Option {
	return func(c *Client) error {
		if endpoint == "" {
			return errors.New("endpoint cannot be empty")
		}
		// Simple URL parse validation
		if _, err := url.ParseRequestURI(endpoint); err != nil {
			return fmt.Errorf("invalid endpoint URL: %w", err)
		}
		c.endpoint = endpoint
		return nil
	}
}

// WithMode sets interactive mode
func WithMode(interactive bool) Option {
	return func(c *Client) error {
		c.interactive = interactive
		return nil
	}
}

// NewClient creates a new client with options applied and validated
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		endpoint:    DefaultEndpoint,
		timeout:     DefaultTimeout,
		interactive: false,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return c, nil
}

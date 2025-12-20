package registry

import (
	"fmt"
	"strings"
	"time"
)

type Option func(*Engine) error

func WithStorageEndpoint(endpoint string) Option {
	return func(e *Engine) error {
		// set endpoint
		if endpoint == "" {
			return fmt.Errorf("storage endpoint value required")
		}
		e.storage = Endpoint(strings.TrimSpace(endpoint))

		return nil
	}
}

func WithBucket(bucket string) Option {
	return func(e *Engine) error {
		// set bucket
		if bucket == "" {
			return fmt.Errorf("storage bucket value required")
		}
		e.bucket = strings.TrimSpace(bucket)

		return nil
	}
}

func WithBucketPrefix(prefix string) Option {
	return func(e *Engine) error {
		e.prefix = strings.TrimSpace(prefix)
		return nil
	}
}

func WithDatabaseEndpoint(endpoint string) Option {
	return func(e *Engine) error {
		if endpoint == "" {
			return fmt.Errorf("database endpoint value required")
		}
		e.database = Endpoint(strings.TrimSpace(endpoint))
		return nil
	}
}

func WithDatabaseUser(user string) Option {
	return func(e *Engine) error {
		if user == "" {
			return fmt.Errorf("database user value required")
		}
		e.databaseUser = strings.TrimSpace(user)
		return nil
	}
}

func WithDatabasePassword(password string) Option {
	return func(e *Engine) error {
		if password == "" {
			return fmt.Errorf("database password value required")
		}
		e.databasePassword = Secret(strings.TrimSpace(password))
		return nil
	}
}

func WithDatabaseName(name string) Option {
	return func(e *Engine) error {
		e.databaseName = strings.TrimSpace(name)
		return nil
	}
}

func WithTimeZone(timeZone string) Option {
	return func(e *Engine) error {
		if _, err := time.LoadLocation(timeZone); err != nil {
			return fmt.Errorf("invalid timezone: %s", timeZone)
		}
		e.timeZone = timeZone
		return nil
	}
}

func WithSslMode() Option {
	return func(e *Engine) error {
		e.databaseSslMode = true
		return nil
	}
}

package registry

import (
	"fmt"
	"strings"

	"github.com/UnivocalX/aether/pkg/errors"
)

type Config struct {
	S3Endpoint string
	Bucket     string
	Prefix     string
	Rich       bool
}

// Validate checks if the configuration is valid
func (cfg *Config) Validate() error {
	var err []string

	if cfg.S3Endpoint == "" {
		err = append(err, "S3Endpoint is required")
	}

	if cfg.Bucket == "" {
		err = append(err, "Bucket is required")
	} else if strings.Contains(cfg.Bucket, " ") {
		err = append(err, "Bucket cannot contain spaces")
	}

	// Prefix is optional, but if provided, should be valid
	if cfg.Prefix != "" {
		if strings.HasPrefix(cfg.Prefix, "/") || strings.HasSuffix(cfg.Prefix, "/") {
			err = append(err, "Prefix should not start or end with '/'")
		}
		if strings.Contains(cfg.Prefix, "//") {
			err = append(err, "Prefix cannot contain consecutive slashes")
		}
	}

	if len(err) > 0 {
		return fmt.Errorf("%w, %s", errors.ErrValidation, strings.Join(err, "; "))
	}

	return nil
}

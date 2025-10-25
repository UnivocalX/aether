package registry

import (
	"fmt"
	"strings"
)

type Config struct {
	S3Endpoint string
	Bucket     string
	Prefix     string
}

// Validate checks if the configuration is valid
func (cfg *Config) Validate() error {
    var errs []string

    if cfg.S3Endpoint == "" {
        errs = append(errs, "S3Endpoint is required")
    }

    if cfg.Bucket == "" {
        errs = append(errs, "Bucket is required")
    } else if strings.Contains(cfg.Bucket, " ") {
        errs = append(errs, "Bucket cannot contain spaces")
    }

    // Prefix is optional, but if provided, should be valid
    if cfg.Prefix != "" {
        if strings.HasPrefix(cfg.Prefix, "/") || strings.HasSuffix(cfg.Prefix, "/") {
            errs = append(errs, "Prefix should not start or end with '/'")
        }
        if strings.Contains(cfg.Prefix, "//") {
            errs = append(errs, "Prefix cannot contain consecutive slashes")
        }
    }

    if len(errs) > 0 {
        return fmt.Errorf("%s", strings.Join(errs, "; "))
    }

    return nil
}

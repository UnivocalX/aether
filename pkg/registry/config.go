package registry

import (
	"fmt"
	"time"
)

type Options struct {
	S3Endpoint string        `json:"endpoint"`
	Bucket     string        `json:"bucket"`
	Prefix     string        `json:"prefix"`
	TTL        time.Duration `json:"ttl"` // Default expiry
}

// Normalize checks configuration
func (o *Options) Normalize() error {
	if o.Bucket == "" {
		return fmt.Errorf("bucket required")
	}

	// Set default TTL
	if o.TTL == 0 {
		o.TTL = 15 * time.Minute
	}

	// Validate TTL bounds
	if o.TTL < time.Minute {
		return fmt.Errorf("TTL too short")
	}
	if o.TTL > 7*24*time.Hour {
		return fmt.Errorf("TTL too long")
	}

	return nil
}

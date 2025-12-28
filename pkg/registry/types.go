package registry

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"database/sql/driver"
)

type Status string

func (p *Status) Scan(value interface{}) error {
	if value == nil {
		*p = StatusPending
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*p = Status(v)
	case string:
		*p = Status(v)
	default:
		return fmt.Errorf("cannot scan type %T into Status", value)
	}

	return nil
}

func (p Status) Value() (driver.Value, error) {
	return string(p), nil
}

const (
	StatusPending  Status = "pending"
	StatusReady    Status = "ready"
	StatusRejected Status = "rejected"
	StatusDeleted  Status = "deleted"
)

// ### Secret Type ###
type Secret string

func (s Secret) String() string {
	return "REDACTED"
}

func (s Secret) Value() string {
	return string(s)
}

// ### DSN Type ###
type DSN string

func (d DSN) String() string {
	// Redact password when converting to string (for logging)
	s := string(d)

	// Simple regex or string replacement to hide password
	return strings.ReplaceAll(s,
		regexp.MustCompile(`password=[^ ]*`).FindString(s),
		"password=REDACTED")
}

// Value returns the actual DSN for database connection
func (d DSN) Value() string {
	return string(d)
}

// ### Endpoint Type ###
type Endpoint string

// GetHost returns the host part of the endpoint (without scheme and port)
func (e Endpoint) GetHost(defaultHost string) string {
	endpoint := strings.TrimSpace(string(e))
	if endpoint == "" {
		return defaultHost
	}

	// Remove scheme if present
	endpoint = removeScheme(endpoint)

	// Handle IPv6 addresses
	if strings.HasPrefix(endpoint, "[") {
		// IPv6: [::1]:8080 or [::1]
		closeBracket := strings.Index(endpoint, "]")
		if closeBracket != -1 {
			return endpoint[1:closeBracket] // Return just the IPv6 address without brackets
		}
		return defaultHost
	}

	// IPv4 or hostname - strip port if present
	lastColon := strings.LastIndex(endpoint, ":")
	if lastColon != -1 {
		// Check if this looks like a port (numeric after colon)
		portPart := endpoint[lastColon+1:]
		if _, err := strconv.Atoi(portPart); err == nil {
			return endpoint[:lastColon]
		}
	}

	// No port found, return the whole endpoint
	return endpoint
}

// GetPort returns the port number from the endpoint
func (e Endpoint) GetPort(defaultPort int) int {
	endpoint := strings.TrimSpace(string(e))
	if endpoint == "" {
		return defaultPort
	}

	// Remove scheme if present
	endpoint = removeScheme(endpoint)

	// Handle IPv6 addresses
	if strings.HasPrefix(endpoint, "[") {
		// IPv6: [::1]:8080
		closeBracket := strings.Index(endpoint, "]")
		if closeBracket != -1 && closeBracket+1 < len(endpoint) && endpoint[closeBracket+1] == ':' {
			portStr := endpoint[closeBracket+2:]
			if port, err := strconv.Atoi(portStr); err == nil && port > 0 && port <= 65535 {
				return port
			}
		}
		return defaultPort
	}

	// IPv4 or hostname - find last colon for port
	lastColon := strings.LastIndex(endpoint, ":")
	if lastColon != -1 {
		portStr := endpoint[lastColon+1:]
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 && port <= 65535 {
			return port
		}
	}

	return defaultPort
}

// GetScheme returns the scheme/protocol from the endpoint
func (e Endpoint) GetScheme(defaultScheme string) string {
	endpoint := strings.TrimSpace(string(e))
	if endpoint == "" {
		return defaultScheme
	}

	// Look for scheme separator
	schemeEnd := strings.Index(endpoint, "://")
	if schemeEnd == -1 {
		return defaultScheme
	}

	scheme := endpoint[:schemeEnd]
	return strings.ToLower(scheme)
}

// Helper function to remove scheme from endpoint
func removeScheme(endpoint string) string {
	schemeEnd := strings.Index(endpoint, "://")
	if schemeEnd != -1 {
		return endpoint[schemeEnd+3:] // Skip "://"
	}
	return endpoint
}

// String returns the endpoint as-is for debugging
func (e Endpoint) String() string {
	return string(e)
}

// PresignedUrl contains presigned URL information including expiry and metadata
type PresignedUrl struct {
	URL       Secret        `json:"url"`
	ExpiresAt time.Time     `json:"expires_at"`
	ExpiresIn time.Duration `json:"expires_in"`
	Checksum  string        `json:"checksum,omitempty"`
	Key       string        `json:"key"`
	Operation string        `json:"operation"`
	Bucket    string        `json:"bucket"`
}

// IsExpired checks if the presigned URL has expired
func (p *PresignedUrl) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// TimeUntilExpiry returns the duration until the URL expires
func (p *PresignedUrl) TimeUntilExpiry() time.Duration {
	return time.Until(p.ExpiresAt)
}

func (p *PresignedUrl) String() string {
	return fmt.Sprintf("PresignUrl{Operation: %s, Bucket: %s, Key: %s, ExpiresAt: %s, ExpiresIn: %s}",
		p.Operation,
		p.Bucket,
		p.Key,
		p.ExpiresAt.Format(time.RFC3339),
		p.ExpiresIn,
	)
}
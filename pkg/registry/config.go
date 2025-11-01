package registry

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Secret string

func (s Secret) String() string {
	return "REDACTED"
}

// Value returns the actual secret value
func (s Secret) Value() string {
	return string(s)
}

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

type Endpoint string

func (e Endpoint) GetHost() string {
	endpoint := strings.TrimSpace(string(e))
	if endpoint == "" {
		return "localhost" // Default host
	}

	parts := strings.Split(endpoint, ":")
	host := strings.TrimSpace(parts[0])
	if host == "" {
		return "localhost"
	}
	return host
}

func (e Endpoint) GetPort() int {
	endpoint := strings.TrimSpace(string(e))
	if endpoint == "" {
		return 5432 // Default PostgreSQL port
	}

	parts := strings.Split(endpoint, ":")
	if len(parts) > 1 {
		portStr := strings.TrimSpace(parts[1])
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 && port <= 65535 {
			return port
		}
	}
	return 5432 // Default PostgreSQL port
}

// Add this method for better debugging
func (e Endpoint) String() string {
	return string(e)
}

// Config is the root configuration for registry
type Config struct {
	Storage  StorageCFG  `json:"storage"`
	Database DatabaseCFG `json:"database"`
}

// DatabaseCFG holds SQL database configuration
type DatabaseCFG struct {
	Endpoint Endpoint `json:"endpoint"`
	User     string   `json:"user"`
	Password Secret   `json:"password"`
	Name     string   `json:"name"`
	SSL      bool     `json:"ssl"`
	TimeZone string   `json:"timezone"`
}

// StorageCFG holds object storage configuration (e.g., S3)
type StorageCFG struct {
	S3Endpoint string        `json:"endpoint"`
	Bucket     string        `json:"bucket"`
	Prefix     string        `json:"prefix"`
	TTL        time.Duration `json:"ttl"`
}

// NewConfig creates a Config with sensible defaults
func NewConfig() *Config {
	return &Config{
		Storage: StorageCFG{
			TTL: 15 * time.Minute,
		},
		Database: DatabaseCFG{
			SSL:      false,
			TimeZone: "UTC",
		},
	}
}

// Validate normalizes and validates the config
func (cfg *Config) Validate() error {
	if err := cfg.Normalize(); err != nil {
		return err
	}

	if err := cfg.Storage.Validate(); err != nil {
		return fmt.Errorf("storage: %w", err)
	}

	if err := cfg.Database.Validate(); err != nil {
		return fmt.Errorf("database: %w", err)
	}

	return nil
}

// Normalize runs Normalize on all sub-configs
func (cfg *Config) Normalize() error {
	if err := cfg.Storage.Normalize(); err != nil {
		return fmt.Errorf("storage: %w", err)
	}
	if err := cfg.Database.Normalize(); err != nil {
		return fmt.Errorf("database: %w", err)
	}
	return nil
}

// --- StorageCFG methods ---

func (cfg *StorageCFG) Normalize() error {
	cfg.S3Endpoint = strings.TrimSpace(cfg.S3Endpoint)
	cfg.Bucket = strings.TrimSpace(cfg.Bucket)
	cfg.Prefix = strings.TrimSpace(cfg.Prefix)

	if cfg.TTL == 0 {
		cfg.TTL = 15 * time.Minute
	}

	return nil
}

func (cfg *StorageCFG) Validate() error {
	if cfg.Bucket == "" {
		return fmt.Errorf("bucket required")
	}
	if cfg.TTL < time.Minute {
		return fmt.Errorf("TTL too short: minimum 1 minute")
	}
	if cfg.TTL > 7*24*time.Hour {
		return fmt.Errorf("TTL too long: maximum 7 days")
	}
	return nil
}

func (cfg StorageCFG) String() string {
	return fmt.Sprintf("StorageCFG{S3Endpoint:%q, Bucket:%q, Prefix:%q, TTL:%s}",
		cfg.S3Endpoint, cfg.Bucket, cfg.Prefix, cfg.TTL)
}

// --- DatabaseCFG methods ---

func (cfg *DatabaseCFG) Normalize() error {
	cfg.User = strings.TrimSpace(cfg.User)
	cfg.Name = strings.TrimSpace(cfg.Name)
	cfg.TimeZone = strings.TrimSpace(cfg.TimeZone)

	if cfg.TimeZone == "" {
		cfg.TimeZone = "UTC"
	}

	// Validate that timezone exists
	if _, err := time.LoadLocation(cfg.TimeZone); err != nil {
		return fmt.Errorf("invalid timezone: %s", cfg.TimeZone)
	}

	return nil
}

func (cfg *DatabaseCFG) Validate() error {
	if cfg.User == "" {
		return fmt.Errorf("user required")
	}
	if cfg.Password == "" {
		return fmt.Errorf("password required")
	}
	if cfg.Name == "" {
		return fmt.Errorf("database name required")
	}
	if cfg.Endpoint.GetHost() == "" {
		return fmt.Errorf("endpoint host required")
	}
	if cfg.Endpoint.GetPort() <= 0 || cfg.Endpoint.GetPort() > 65535 {
		return fmt.Errorf("invalid endpoint port: %d", cfg.Endpoint.GetPort())
	}
	return nil
}

func (cfg *DatabaseCFG) DSN() DSN {
	sslMode := "disable"
	if cfg.SSL {
		sslMode = "require"
	}

	host := cfg.Endpoint.GetHost()
	
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		host, cfg.Endpoint.GetPort(),
		cfg.User, cfg.Password.Value(), cfg.Name, sslMode, cfg.TimeZone,
	)

	return DSN(dsn)
}

// String redacts sensitive fields when printing
func (cfg DatabaseCFG) String() string {
	return fmt.Sprintf(
		"DatabaseCFG{Endpoint:%s, User:%q, Name:%q, SSL:%t, TimeZone:%q, Password:%q}",
		cfg.Endpoint, cfg.User, cfg.Name, cfg.SSL, cfg.TimeZone, cfg.Password,
	)
}

func (cfg Config) String() string {
	return fmt.Sprintf(
		"RegistryConfig{Storage:%s, Database:%s}",
		cfg.Storage, cfg.Database,
	)
}

package registry

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Endpoint string

func (e Endpoint) GetHost() string {
	parts := strings.Split(string(e), ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func (e Endpoint) GetPort() int {
	parts := strings.Split(string(e), ":")
	if len(parts) > 1 {
		port, err := strconv.Atoi(parts[1])
		if err == nil {
			return port
		}
	}
	return 0
}

// Config is the root configuration for aether
type Config struct {
	Storage   StorageCFG   `json:"storage"`
	Datastore DatastoreCFG `json:"datastore"`
}

// DatastoreCFG holds SQL datastore configuration
type DatastoreCFG struct {
	Endpoint Endpoint `json:"endpoint"`
	User     string   `json:"user"`
	Password string   `json:"password"`
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
		Datastore: DatastoreCFG{
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

	if err := cfg.Datastore.Validate(); err != nil {
		return fmt.Errorf("datastore: %w", err)
	}

	return nil
}

// Normalize runs Normalize on all sub-configs
func (cfg *Config) Normalize() error {
	if err := cfg.Storage.Normalize(); err != nil {
		return fmt.Errorf("storage: %w", err)
	}
	if err := cfg.Datastore.Normalize(); err != nil {
		return fmt.Errorf("datastore: %w", err)
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

// --- DatastoreCFG methods ---

func (cfg *DatastoreCFG) Normalize() error {
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

func (cfg *DatastoreCFG) Validate() error {
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

func (cfg *DatastoreCFG) DSN() string {
	sslMode := "disable"
	if cfg.SSL {
		sslMode = "require"
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Endpoint.GetHost(), cfg.Endpoint.GetPort(),
		cfg.User, cfg.Password, cfg.Name, sslMode, cfg.TimeZone,
	)
}

// URL returns the connection URL form (useful for pgx or other drivers)
func (cfg *DatastoreCFG) URL() string {
	sslMode := "disable"
	if cfg.SSL {
		sslMode = "require"
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&TimeZone=%s",
		cfg.User, cfg.Password,
		cfg.Endpoint.GetHost(), cfg.Endpoint.GetPort(),
		cfg.Name, sslMode, cfg.TimeZone,
	)
}

// String redacts sensitive fields when printing
func (cfg DatastoreCFG) String() string {
	return fmt.Sprintf(
		"DatastoreCFG{Endpoint:%s, User:%q, Name:%q, SSL:%t, TimeZone:%q, Password:REDACTED}",
		cfg.Endpoint, cfg.User, cfg.Name, cfg.SSL, cfg.TimeZone,
	)
}

func (cfg Config) String() string {
	return fmt.Sprintf(
		"AetherConfig{Storage:%s, Datastore:%s}",
		cfg.Storage, cfg.Datastore,
	)
}

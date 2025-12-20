package registry

import (
	"fmt"
)

type SearchAssetsConfig struct {
	Cursor       uint
	Limit        uint
	MimeType     string
	State        Status
	IncludedTags []string
	ExcludedTags []string
}

func NewSearchAssetsOptions() SearchAssetsConfig {
	return SearchAssetsConfig{
		Cursor:       0,
		Limit:        150,
		MimeType:     "",
		State:        "",
		IncludedTags: []string{},
		ExcludedTags: []string{},
	}
}

// Validate checks if the search options are valid
func (cfg *SearchAssetsConfig) Validate() error {
	// Set default limit
	if cfg.Limit == 0 {
		cfg.Limit = 150
	}

	// Prevent excessive queries
	if cfg.Limit > 1000 {
		cfg.Limit = 1000
	}

	// Validate MIME type format if provided
	if cfg.MimeType != "" && !ValidateString(cfg.MimeType) {
		return fmt.Errorf("invalid MIME type format: %s", cfg.MimeType)
	}

	// Validate tag names
	for _, tag := range append(cfg.IncludedTags, cfg.ExcludedTags...) {
		if !ValidateString(tag) {
			return fmt.Errorf("invalid tag name: %s", tag)
		}
	}

	return nil
}

// Normalize applies normalization to all fields
func (cfg *SearchAssetsConfig) Normalize() {
	cfg.MimeType = NormalizeString(cfg.MimeType)
	cfg.IncludedTags = normalizeTags(cfg.IncludedTags)
	cfg.ExcludedTags = normalizeTags(cfg.ExcludedTags)
}

// normalizeTags cleans up tag arrays using existing functions
func normalizeTags(tags []string) []string {
	if len(tags) == 0 {
		return tags
	}

	seen := make(map[string]bool)
	result := make([]string, 0, len(tags))

	for _, tag := range tags {
		normalized := NormalizeString(tag)
		if normalized != "" && !seen[normalized] {
			seen[normalized] = true
			result = append(result, normalized)
		}
	}

	return result
}

// IsEmpty checks if this is effectively an empty search
func (cfg *SearchAssetsConfig) IsEmpty() bool {
	return cfg.Cursor == 0 &&
		cfg.MimeType == "" &&
		cfg.State == "" &&
		len(cfg.IncludedTags) == 0 &&
		len(cfg.ExcludedTags) == 0
}

package registry

import (
	"fmt"
)

const (
	SearchDefaultLimit = 150
	SearchMaxLimit     = 1000
)

type SearchAssetsQuery struct {
	Cursor       uint
	Limit        uint
	MimeType     string
	State        Status
	IncludedTags []string
	ExcludedTags []string
	CheckSums    []string
}

func (q SearchAssetsQuery) String() string {
	return fmt.Sprintf(
		"SearchAssetsQuery{Cursor: %d, Limit: %d, MimeType: %q, State: %v, IncludedTags: %v, ExcludedTags: %v}",
		q.Cursor, q.Limit, q.MimeType, q.State, q.IncludedTags, q.ExcludedTags,
	)
}

type SearchAssetsOption func(*SearchAssetsQuery) error

func WithCursor(cursor uint) SearchAssetsOption {
	return func(q *SearchAssetsQuery) error {
		q.Cursor = cursor
		return nil
	}
}

func WithLimit(limit uint) SearchAssetsOption {
	return func(q *SearchAssetsQuery) error {
		if limit == 0 {
			return fmt.Errorf("limit must be greater than 0")
		}
		if limit > SearchMaxLimit {
			return fmt.Errorf("limit cannot exceed %d", SearchMaxLimit)
		}
		q.Limit = limit
		return nil
	}
}

func WithMimeType(mimeType string) SearchAssetsOption {
	return func(q *SearchAssetsQuery) error {
		normalized := NormalizeString(mimeType)
		if normalized == "" {
			return fmt.Errorf("mime type cannot be empty")
		}

		q.MimeType = normalized
		return nil
	}
}

func WithState(state Status) SearchAssetsOption {
	return func(q *SearchAssetsQuery) error {
		q.State = state
		return nil
	}
}

func WithIncludedTags(tags ...string) SearchAssetsOption {
	return func(q *SearchAssetsQuery) error {
		if len(tags) == 0 {
			return fmt.Errorf("at least one included tag must be provided")
		}

		cleaned := make([]string, 0, len(tags))
		seen := make(map[string]bool)

		for _, tag := range tags {
			normalized := NormalizeString(tag)
			if normalized == "" {
				return fmt.Errorf("tag cannot be empty or whitespace")
			}

			if seen[normalized] {
				continue
			}

			seen[normalized] = true
			cleaned = append(cleaned, normalized)
		}
		q.IncludedTags = cleaned
		return nil
	}
}

func WithExcludedTags(tags ...string) SearchAssetsOption {
	return func(q *SearchAssetsQuery) error {
		if len(tags) == 0 {
			return fmt.Errorf("at least one excluded tag must be provided")
		}

		cleaned := make([]string, 0, len(tags))
		seen := make(map[string]bool)

		for _, tag := range tags {
			normalized := NormalizeString(tag)
			if normalized == "" {
				return fmt.Errorf("tag cannot be empty or whitespace")
			}

			if seen[normalized] {
				continue
			}

			seen[normalized] = true
			cleaned = append(cleaned, normalized)
		}
		q.ExcludedTags = cleaned
		return nil
	}
}

func WithChecksums(checksums ...string) SearchAssetsOption {
	return func(q *SearchAssetsQuery) error {
		if len(checksums) == 0 {
			return fmt.Errorf("at least one excluded checksum must be provided")
		}

		q.CheckSums = checksums
		return nil
	}
}

func NewSearchAssetsQuery(opts ...SearchAssetsOption) (*SearchAssetsQuery, error) {
	// Initialize with defaults
	query := &SearchAssetsQuery{
		Cursor:       0,
		Limit:        SearchDefaultLimit,
		IncludedTags: []string{},
		ExcludedTags: []string{},
	}

	// Apply all options
	for _, opt := range opts {
		if err := opt(query); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return query, nil
}

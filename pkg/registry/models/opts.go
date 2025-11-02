package models

import (
	"fmt"
)

type SearchAssetsOptions struct {
    Cursor         uint
    Limit          uint
    MimeType       string
    State          Status
    IncludedTags   []string
    ExcludedTags   []string
}

func NewSearchAssetsOptions() SearchAssetsOptions {
    return SearchAssetsOptions{
        Cursor:       0,
        Limit:        150,
        MimeType:     "",
        State:        "",
        IncludedTags: []string{},
        ExcludedTags: []string{},
    }
}

// Validate checks if the search options are valid
func (opt *SearchAssetsOptions) Validate() error {    
    // Set default limit
    if opt.Limit == 0 {
        opt.Limit = 150
    }
    
    // Prevent excessive queries
    if opt.Limit > 1000 {
        opt.Limit = 1000
    }
    
    // Validate MIME type format if provided
    if opt.MimeType != "" && !ValidateName(opt.MimeType) {
        return fmt.Errorf("invalid MIME type format: %s", opt.MimeType)
    }
    
    // Validate tag names
    for _, tag := range append(opt.IncludedTags, opt.ExcludedTags...) {
        if !ValidateName(tag) {
            return fmt.Errorf("invalid tag name: %s", tag)
        }
    }
    
    return nil
}

// Normalize applies normalization to all fields
func (opt *SearchAssetsOptions) Normalize() {
    opt.MimeType = NormalizeName(opt.MimeType)
    opt.IncludedTags = normalizeTags(opt.IncludedTags)
    opt.ExcludedTags = normalizeTags(opt.ExcludedTags)
}

// normalizeTags cleans up tag arrays using existing functions
func normalizeTags(tags []string) []string {
    if len(tags) == 0 {
        return tags
    }
    
    seen := make(map[string]bool)
    result := make([]string, 0, len(tags))
    
    for _, tag := range tags {
        normalized := NormalizeName(tag)
        if normalized != "" && !seen[normalized] {
            seen[normalized] = true
            result = append(result, normalized)
        }
    }
    
    return result
}

// IsEmpty checks if this is effectively an empty search
func (opt *SearchAssetsOptions) IsEmpty() bool {
    return opt.Cursor == 0 && 
           opt.MimeType == "" && 
           opt.State == "" && 
           len(opt.IncludedTags) == 0 && 
           len(opt.ExcludedTags) == 0
}
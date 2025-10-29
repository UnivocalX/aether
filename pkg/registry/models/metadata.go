package models

import (
	"fmt"
	"strings"
)

type AssetMetadata struct {
	MimeType  string                 `json:"mime_type" validate:"required"`
	SizeBytes int64                  `json:"size_bytes" validate:"required,gt=0"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

func (m *AssetMetadata) Validate() error {
	if m.MimeType == "" {
		return fmt.Errorf("mime_type is required")
	}
	if !strings.Contains(m.MimeType, "/") {
		return fmt.Errorf("mime_type must be in format 'type/subtype'")
	}
	if m.SizeBytes <= 0 {
		return fmt.Errorf("size_bytes must be greater than 0")
	}
	return nil
}

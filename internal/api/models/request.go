package models

import (
	"strings"

	"github.com/UnivocalX/aether/internal/utils"
)

// CreateDataRequest contains the JSON body and URI path parameter for POST /data/:sha256.
type CreateDataRequest struct {
	SHA256 string `uri:"sha256" binding:"required,len=64,hexadecimal"`
	Tag    string `json:"tag,omitempty"`
}

// Normalize trims + lowercases the SHA and then validates it.
func (r *CreateDataRequest) Normalize() error {
	r.SHA256 = strings.ToLower(strings.TrimSpace(r.SHA256))
	return utils.ValidateSHA256(r.SHA256)
}
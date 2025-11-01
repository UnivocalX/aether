package schemas

import (
	"github.com/UnivocalX/aether/pkg/registry/models"
)

// Standard API Response - Minimalist version
type Response struct {
	Message string            `json:"message"`
	Data    interface{}       `json:"data,omitempty"`
	Meta    *ResponseMetadata `json:"meta,omitempty"`
}

type ResponseMetadata struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id,omitempty"`
}

type CreateAssetResponse struct {
	AssetID      uint   `json:"asset_id"`
	SHA256       string `json:"sha256"`
	PresignedURL string `json:"presigned_url"`
	Expiry       string `json:"expiry"`
}

type CreateTagResponse struct {
	Name string `json:"name"`
	ID   uint   `json:"id"`
}

type GetTagResponse struct {
	Tag *models.Tag `json:",inline"`
}

type GetAssetResponse struct {
	Asset *models.Asset `json:",inline"`
}

type ListTagsResponse struct {
	Tags       []*models.Tag `json:"tags"`
	NextCursor uint          `json:"next_cursor,omitempty"`
	HasMore    bool          `json:"has_more"`
}

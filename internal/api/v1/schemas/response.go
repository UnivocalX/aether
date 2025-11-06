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

type PaginatedResponse struct {
	NextCursor uint `json:"nextCursor"`
	HasMore    bool `json:"hasMore"`
}

type ResponseMetadata struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id,omitempty"`
}

type BatchCreateAssetResponse struct {
	CreateAssetResponse
	Error string `json:"error,omitempty"`
}

type CreateAssetResponse struct {
	AssetID      uint            `json:"asset_id"`
	SHA256       string          `json:"sha256"`
	PresignedURL string `json:"presigned_url"`
	Expiry       string          `json:"expiry"`
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
	PaginatedResponse
	Tags []*models.Tag `json:"tags"`
}

type ListAssetsResponse struct {
	PaginatedResponse
	Assets []*models.Asset `json:"assets"`
}

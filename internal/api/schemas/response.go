package schemas

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
	SHA256       string `json:"sha256"`
	PresignedURL string `json:"presigned_url"`
	Expiry       string `json:"expiry"`
	AssetID      uint   `json:"asset_id"`
}

type CreateTagResponse struct {
	Name string `json:"name"`
	ID   uint   `json:"id"`
}

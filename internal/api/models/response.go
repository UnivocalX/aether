package models

// Standard API Response - Minimalist version
type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Metadata   `json:"meta,omitempty"`
}

type Metadata struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id,omitempty"`
}

type CreateResponseData struct {
	SHA256       string `json:"sha256"`
	PresignedURL string `json:"presigned_url"`
	Expiry       string `json:"expiry"` 
}

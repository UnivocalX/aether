package schemas

import (
	"github.com/UnivocalX/aether/pkg/registry/models"
)

type BatchCreateAssetRequest struct {
	Assets []CreateAssetRequest `json:"assets" binding:"required,min=1,max=900"`
	Tags   []uint               `json:"tags" binding:"omitempty,dive,gt=0"` // Global tags for all assets
}

type CreateAssetRequest struct {
	SHA256  string                 `uri:"sha256" binding:"required,len=64,hexadecimal"`
	Display string                 `json:"display" binding:"required,min=1,max=500"`
	Tags    []uint                 `json:"tags" binding:"omitempty,dive,gt=0"`
	Extra   map[string]interface{} `json:"extra" binding:"omitempty"`
}

type CreateTagRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type GetTagRequest struct {
	Name string `uri:"name" binding:"required,min=1,max=100"`
}

type GetAssetRequest struct {
	SHA256 string `uri:"sha256" binding:"required,len=64,hexadecimal"`
}

type ListTagsRequest struct {
	Cursor uint `form:"cursor" binding:"omitempty,min=0"`
	Limit  int  `form:"limit" binding:"omitempty,min=1,max=100"`
}

type ListAssetsRequest struct {
	models.SearchAssetsOptions
}

package v1

type AssetListGetPayload struct {
	Display string                 `json:"display" binding:"omitempty,max=500"`
	Tags    []uint                 `json:"tags" binding:"omitempty,dive,gt=0"`
	Extra   map[string]interface{} `json:"extra" binding:"omitempty"`
}

type AssetListGetRequest struct {
	AssetListGetPayload
}

package schemas

type CreateAssetRequest struct {
	SHA256  string `uri:"sha256" binding:"required,len=64,hexadecimal"`
	Display string `json:"display" binding:"required,min=1,max=500"`
	Tags    []uint `json:"tags" binding:"omitempty,dive,gt=0"`
}

type CreateTagRequest struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

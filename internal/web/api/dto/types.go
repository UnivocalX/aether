package dto

type TagUri struct {
	TagName string `uri:"tag_name" binding:"required,min=1,max=100"`
}

type AssetUri struct {
	AssetChecksum string `uri:"asset_checksum" binding:"required,len=64,hexadecimal"`
}

type DatasetUri struct {
	DatasetName string `uri:"dataset_name" binding:"required,max=100"`
}

type AssetTagUri struct {
	TagUri
	AssetUri
}

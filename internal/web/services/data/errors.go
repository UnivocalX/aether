package data

import (
	"errors"
	"fmt"
)

var (
	ErrAssetNotFound        = errors.New("asset not found")
	ErrTagNotFound          = errors.New("tag not found")
	ErrAssetAlreadyExists   = errors.New("asset already exists")
	ErrTagAlreadyExists     = errors.New("tag already exists")
	ErrDatasetAlreadyExists = errors.New("dataset already exists")
	ErrAssetIsReady         = errors.New("reuploading a ready asset is not allowed")
)

type ErrAssetsExists struct {
	Checksums []string
}

func (e ErrAssetsExists) Error() string {
	return fmt.Sprintf("%d assets already exist", len(e.Checksums))
}


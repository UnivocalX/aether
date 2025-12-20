package data

import (
	"errors"
)

var (
	ErrAssetNotFound = errors.New("asset not found")
	ErrTagNotFound   = errors.New("tag not found")
	ErrAssetAlreadyExists = errors.New("asset already exists")
	ErrTagAlreadyExists = errors.New("tag already exists")
)
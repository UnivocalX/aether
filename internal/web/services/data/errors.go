package data

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrAssetNotFound            = errors.New("asset not found")
	ErrTagNotFound              = errors.New("tag not found")
	ErrAssetAlreadyExists       = errors.New("asset already exists")
	ErrTagAlreadyExists         = errors.New("tag already exists")
	ErrDatasetAlreadyExists     = errors.New("dataset already exists")
	ErrCantGeneratePresignedUrl = errors.New("cant generate presigned url")
	ErrAssetIsReady             = errors.New("reuploading a ready asset is not allowed")
)

type MultiError struct {
	Errors []*error
}

func (e MultiError) Error() string {
	return fmt.Sprintf("%d errors occurred", len(e.Errors))
}

type AssetsExistsError struct {
	Checksums []*string
}

func (e AssetsExistsError) Error() string {
	return fmt.Sprintf("%d asset(s) already exist", len(e.Checksums))
}

func IsUniqueConstraintError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

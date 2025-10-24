package errors

import (
	"errors"
)

var (
	ErrValue      = errors.New("value error")
	ErrValidation = errors.New("validation error")
)

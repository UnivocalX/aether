package dto

import "errors"


var (
	ErrInvalidUri = errors.New("Invalid URI parameters")
	ErrInvalidPayload = errors.New("Invalid payload")
)
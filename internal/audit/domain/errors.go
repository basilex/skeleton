package domain

import "errors"

var (
	ErrRecordNotFound = errors.New("audit record not found")
	ErrInvalidFilter  = errors.New("invalid audit filter")
)

package catalog

import "errors"

var (
	ErrItemNotFound          = errors.New("item not found")
	ErrItemAlreadyExists     = errors.New("item already exists")
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrInvalidSKU            = errors.New("invalid SKU")
	ErrInvalidPrice          = errors.New("invalid price")
	ErrItemInactive          = errors.New("item is inactive")
	ErrItemDiscontinued      = errors.New("item is discontinued")
)

package domain

import (
	"errors"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrInvalidPassword        = errors.New("invalid password")
	ErrUserInactive           = errors.New("user is inactive")
	ErrRoleNotFound           = errors.New("role not found")
	ErrRoleAlreadyExists      = errors.New("role already exists")
	ErrRoleAlreadyAssigned    = errors.New("role already assigned")
	ErrRoleNotAssigned        = errors.New("role not assigned")
	ErrInvalidPermission      = errors.New("invalid permission format")
	ErrInsufficientPermission = errors.New("insufficient permission")
)

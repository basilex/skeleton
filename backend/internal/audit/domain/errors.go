// Package domain provides domain entities and repository interfaces for the audit module.
// This package contains the core business logic types for audit trail tracking and
// repository contracts for persisting audit records.
package domain

import "errors"

var (
	ErrRecordNotFound  = errors.New("audit record not found")
	ErrInvalidFilter   = errors.New("invalid audit filter")
	ErrInvalidRecordID = errors.New("invalid record id")
)

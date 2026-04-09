// Package domain provides domain entities and repository interfaces for the audit module.
// This package contains the core business logic types for audit trail tracking and
// repository contracts for persisting audit records.
package domain

import (
	"context"
	"time"

	"github.com/basilex/skeleton/pkg/pagination"
)

// AuditRepository defines the contract for audit record persistence operations.
type AuditRepository interface {
	Save(ctx context.Context, record *Record) error
	FindAll(ctx context.Context, filter RecordFilter) (pagination.PageResult[*Record], error)
	FindByActorID(ctx context.Context, actorID string, filter RecordFilter) (pagination.PageResult[*Record], error)
	DeleteBefore(ctx context.Context, before time.Time) (int, error)
}

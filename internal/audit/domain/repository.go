package domain

import (
	"context"
	"time"

	"github.com/basilex/skeleton/pkg/pagination"
)

type AuditRepository interface {
	Save(ctx context.Context, record *Record) error
	FindAll(ctx context.Context, filter RecordFilter) (pagination.PageResult[*Record], error)
	FindByActorID(ctx context.Context, actorID string, filter RecordFilter) (pagination.PageResult[*Record], error)
	DeleteBefore(ctx context.Context, before time.Time) (int, error)
}

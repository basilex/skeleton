// Package persistence provides database repository implementations for the audit domain.
// This package contains PostgreSQL-based repositories for audit records.
package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/audit/domain"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

func (r *AuditRepository) Save(ctx context.Context, record *domain.Record) error {
	query := `
		INSERT INTO audit_records (id, actor_id, actor_type, action, resource, resource_id, metadata, ip, user_agent, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		record.ID().String(),
		record.ActorID(),
		record.ActorType().String(),
		record.Action().String(),
		record.Resource(),
		record.ResourceID(),
		record.Metadata(),
		record.IP(),
		record.UserAgent(),
		record.Status(),
		record.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("save audit record: %w", err)
	}
	return nil
}

func (r *AuditRepository) FindAll(ctx context.Context, filter domain.RecordFilter) (pagination.PageResult[*domain.Record], error) {
	query := `SELECT id, actor_id, actor_type, action, resource, resource_id, metadata, ip, user_agent, status, created_at FROM audit_records`
	args := make([]interface{}, 0)
	conditions := make([]string, 0)

	if filter.ActorID != "" {
		conditions = append(conditions, fmt.Sprintf("actor_id = $%d", len(args)+1))
		args = append(args, filter.ActorID)
	}
	if filter.Resource != "" {
		conditions = append(conditions, fmt.Sprintf("resource = $%d", len(args)+1))
		args = append(args, filter.Resource)
	}
	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", len(args)+1))
		args = append(args, filter.Action)
	}
	if !filter.DateFrom.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", len(args)+1))
		args = append(args, filter.DateFrom)
	}
	if !filter.DateTo.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", len(args)+1))
		args = append(args, filter.DateTo)
	}
	if filter.Cursor != "" {
		conditions = append(conditions, fmt.Sprintf("id < $%d", len(args)+1))
		args = append(args, filter.Cursor)
	}

	if len(conditions) > 0 {
		where := " WHERE " + joinConditions(conditions)
		query += where
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = pagination.DefaultLimit
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d`, len(args)+1)
	args = append(args, limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return pagination.PageResult[*domain.Record]{}, fmt.Errorf("select audit records: %w", err)
	}
	defer rows.Close()

	records := make([]*domain.Record, 0)
	for rows.Next() {
		var id, actorID, actorType, action, resource, resourceID, metadata, ip, userAgent string
		var status int
		var createdAt time.Time
		if err := rows.Scan(&id, &actorID, &actorType, &action, &resource, &resourceID, &metadata, &ip, &userAgent, &status, &createdAt); err != nil {
			return pagination.PageResult[*domain.Record]{}, fmt.Errorf("scan record: %w", err)
		}
		recordID, parseErr := domain.ParseRecordID(id)
		if parseErr != nil {
			return pagination.PageResult[*domain.Record]{}, fmt.Errorf("parse record id: %w", parseErr)
		}
		record := domain.ReconstituteRecord(
			recordID,
			actorID,
			domain.ActorType(actorType),
			domain.Action(action),
			resource,
			resourceID,
			metadata,
			ip,
			userAgent,
			status,
			createdAt,
		)
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return pagination.PageResult[*domain.Record]{}, fmt.Errorf("iterate records: %w", err)
	}

	return pagination.NewPageResult(records, limit), nil
}

func (r *AuditRepository) FindByActorID(ctx context.Context, actorID string, filter domain.RecordFilter) (pagination.PageResult[*domain.Record], error) {
	query := `SELECT id, actor_id, actor_type, action, resource, resource_id, metadata, ip, user_agent, status, created_at FROM audit_records WHERE actor_id = $1`
	args := []interface{}{actorID}
	conditions := make([]string, 0)

	if filter.Resource != "" {
		conditions = append(conditions, fmt.Sprintf("resource = $%d", len(args)+1))
		args = append(args, filter.Resource)
	}
	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", len(args)+1))
		args = append(args, filter.Action)
	}
	if !filter.DateFrom.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", len(args)+1))
		args = append(args, filter.DateFrom)
	}
	if !filter.DateTo.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", len(args)+1))
		args = append(args, filter.DateTo)
	}
	if filter.Cursor != "" {
		conditions = append(conditions, fmt.Sprintf("id < $%d", len(args)+1))
		args = append(args, filter.Cursor)
	}

	if len(conditions) > 0 {
		where := " AND " + joinConditions(conditions)
		query += where
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = pagination.DefaultLimit
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d`, len(args)+1)
	args = append(args, limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return pagination.PageResult[*domain.Record]{}, fmt.Errorf("select audit records by actor: %w", err)
	}
	defer rows.Close()

	records := make([]*domain.Record, 0)
	for rows.Next() {
		var id, actorID, actorType, action, resource, resourceID, metadata, ip, userAgent string
		var status int
		var createdAt time.Time
		if err := rows.Scan(&id, &actorID, &actorType, &action, &resource, &resourceID, &metadata, &ip, &userAgent, &status, &createdAt); err != nil {
			return pagination.PageResult[*domain.Record]{}, fmt.Errorf("scan record: %w", err)
		}
		recordID, parseErr := domain.ParseRecordID(id)
		if parseErr != nil {
			return pagination.PageResult[*domain.Record]{}, fmt.Errorf("parse record id: %w", parseErr)
		}
		record := domain.ReconstituteRecord(
			recordID,
			actorID,
			domain.ActorType(actorType),
			domain.Action(action),
			resource,
			resourceID,
			metadata,
			ip,
			userAgent,
			status,
			createdAt,
		)
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return pagination.PageResult[*domain.Record]{}, fmt.Errorf("iterate records: %w", err)
	}

	return pagination.NewPageResult(records, limit), nil
}

func (r *AuditRepository) DeleteBefore(ctx context.Context, before time.Time) (int, error) {
	result, err := r.pool.Exec(ctx, `DELETE FROM audit_records WHERE created_at < $1`, before)
	if err != nil {
		return 0, fmt.Errorf("delete old records: %w", err)
	}
	return int(result.RowsAffected()), nil
}

func joinConditions(conditions []string) string {
	result := ""
	for i, c := range conditions {
		if i > 0 {
			result += " AND "
		}
		result += c
	}
	return result
}

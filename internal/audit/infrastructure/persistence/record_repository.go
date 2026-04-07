// Package persistence provides database repository implementations for the audit domain.
// This package contains SQLite-based repositories for audit records.
package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/audit/domain"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/jmoiron/sqlx"
)

// AuditRepository implements the audit record repository interface using SQL database storage.
// It handles persistence and querying of audit trail records.
type AuditRepository struct {
	db *sqlx.DB
}

// NewAuditRepository creates a new audit repository with the provided database connection.
func NewAuditRepository(db *sqlx.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Save persists an audit record to the database.
func (r *AuditRepository) Save(ctx context.Context, record *domain.Record) error {
	query := `
		INSERT INTO audit_records (id, actor_id, actor_type, action, resource, resource_id, metadata, ip, user_agent, status, created_at)
		VALUES (:id, :actor_id, :actor_type, :action, :resource, :resource_id, :metadata, :ip, :user_agent, :status, :created_at)
	`
	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":          string(record.ID()),
		"actor_id":    record.ActorID(),
		"actor_type":  record.ActorType().String(),
		"action":      record.Action().String(),
		"resource":    record.Resource(),
		"resource_id": record.ResourceID(),
		"metadata":    record.Metadata(),
		"ip":          record.IP(),
		"user_agent":  record.UserAgent(),
		"status":      record.Status(),
		"created_at":  record.CreatedAt().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("save audit record: %w", err)
	}
	return nil
}

// FindAll retrieves audit records with optional filtering and pagination.
// Supports filtering by actor, resource, action, date range, and cursor-based pagination.
func (r *AuditRepository) FindAll(ctx context.Context, filter domain.RecordFilter) (pagination.PageResult[*domain.Record], error) {
	query := `SELECT id, actor_id, actor_type, action, resource, resource_id, metadata, ip, user_agent, status, created_at FROM audit_records`
	args := make([]interface{}, 0)
	conditions := make([]string, 0)

	if filter.ActorID != "" {
		conditions = append(conditions, "actor_id = ?")
		args = append(args, filter.ActorID)
	}
	if filter.Resource != "" {
		conditions = append(conditions, "resource = ?")
		args = append(args, filter.Resource)
	}
	if filter.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, filter.Action)
	}
	if !filter.DateFrom.IsZero() {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, filter.DateFrom.Format(time.RFC3339))
	}
	if !filter.DateTo.IsZero() {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, filter.DateTo.Format(time.RFC3339))
	}
	if filter.Cursor != "" {
		conditions = append(conditions, "id < ?")
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

	query += ` ORDER BY created_at DESC LIMIT ?`
	args = append(args, limit+1)

	var rows []struct {
		ID         string `db:"id"`
		ActorID    string `db:"actor_id"`
		ActorType  string `db:"actor_type"`
		Action     string `db:"action"`
		Resource   string `db:"resource"`
		ResourceID string `db:"resource_id"`
		Metadata   string `db:"metadata"`
		IP         string `db:"ip"`
		UserAgent  string `db:"user_agent"`
		Status     int    `db:"status"`
		CreatedAt  string `db:"created_at"`
	}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return pagination.PageResult[*domain.Record]{}, fmt.Errorf("select audit records: %w", err)
	}

	records := make([]*domain.Record, len(rows))
	for i, row := range rows {
		r, err := r.scanRecord(row)
		if err != nil {
			return pagination.PageResult[*domain.Record]{}, fmt.Errorf("scan record: %w", err)
		}
		records[i] = r
	}

	return pagination.NewPageResult(records, limit), nil
}

// FindByActorID retrieves audit records for a specific actor with optional filtering.
// Returns records ordered by creation date descending.
func (r *AuditRepository) FindByActorID(ctx context.Context, actorID string, filter domain.RecordFilter) (pagination.PageResult[*domain.Record], error) {
	query := `SELECT id, actor_id, actor_type, action, resource, resource_id, metadata, ip, user_agent, status, created_at FROM audit_records WHERE actor_id = ?`
	args := []interface{}{actorID}
	conditions := make([]string, 0)

	if filter.Resource != "" {
		conditions = append(conditions, "resource = ?")
		args = append(args, filter.Resource)
	}
	if filter.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, filter.Action)
	}
	if !filter.DateFrom.IsZero() {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, filter.DateFrom.Format(time.RFC3339))
	}
	if !filter.DateTo.IsZero() {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, filter.DateTo.Format(time.RFC3339))
	}
	if filter.Cursor != "" {
		conditions = append(conditions, "id < ?")
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

	query += ` ORDER BY created_at DESC LIMIT ?`
	args = append(args, limit+1)

	var rows []struct {
		ID         string `db:"id"`
		ActorID    string `db:"actor_id"`
		ActorType  string `db:"actor_type"`
		Action     string `db:"action"`
		Resource   string `db:"resource"`
		ResourceID string `db:"resource_id"`
		Metadata   string `db:"metadata"`
		IP         string `db:"ip"`
		UserAgent  string `db:"user_agent"`
		Status     int    `db:"status"`
		CreatedAt  string `db:"created_at"`
	}
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return pagination.PageResult[*domain.Record]{}, fmt.Errorf("select audit records by actor: %w", err)
	}

	records := make([]*domain.Record, len(rows))
	for i, row := range rows {
		r, err := r.scanRecord(row)
		if err != nil {
			return pagination.PageResult[*domain.Record]{}, fmt.Errorf("scan record: %w", err)
		}
		records[i] = r
	}

	return pagination.NewPageResult(records, limit), nil
}

// DeleteBefore removes audit records older than the specified time.
// Returns the number of records deleted.
func (r *AuditRepository) DeleteBefore(ctx context.Context, before time.Time) (int, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM audit_records WHERE created_at < ?`, before.Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("delete old records: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("rows affected: %w", err)
	}
	return int(affected), nil
}

// scanRecord converts a database row into a domain Record entity.
func (r *AuditRepository) scanRecord(row struct {
	ID         string `db:"id"`
	ActorID    string `db:"actor_id"`
	ActorType  string `db:"actor_type"`
	Action     string `db:"action"`
	Resource   string `db:"resource"`
	ResourceID string `db:"resource_id"`
	Metadata   string `db:"metadata"`
	IP         string `db:"ip"`
	UserAgent  string `db:"user_agent"`
	Status     int    `db:"status"`
	CreatedAt  string `db:"created_at"`
}) (*domain.Record, error) {
	createdAt, err := time.Parse(time.RFC3339, row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	record := domain.ReconstituteRecord(
		domain.RecordID(row.ID),
		row.ActorID,
		domain.ActorType(row.ActorType),
		domain.Action(row.Action),
		row.Resource,
		row.ResourceID,
		row.Metadata,
		row.IP,
		row.UserAgent,
		row.Status,
		createdAt,
	)
	return record, nil
}

// joinConditions concatenates multiple SQL conditions with AND operators.
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

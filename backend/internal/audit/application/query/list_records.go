// Package query provides query handlers for reading audit records.
// This package implements the query side of CQRS for audit-related operations,
// handling read-only requests that return audit record data transfer objects.
package query

import (
	"context"
	"time"

	"github.com/basilex/skeleton/internal/audit/domain"
	"github.com/basilex/skeleton/pkg/pagination"
)

// ListRecordsHandler handles queries to retrieve a paginated list of audit records.
// It supports filtering by actor, resource, action, and date range.
type ListRecordsHandler struct {
	repo domain.AuditRepository
}

// NewListRecordsHandler creates a new ListRecordsHandler with the required repository.
func NewListRecordsHandler(repo domain.AuditRepository) *ListRecordsHandler {
	return &ListRecordsHandler{
		repo: repo,
	}
}

// ListRecordsQuery represents a query to list audit records with optional filtering.
type ListRecordsQuery struct {
	ActorID              string
	Resource             string
	Action               string
	DateFrom             time.Time
	DateTo               time.Time
	Cursor               string
	Limit                int
	RequestedByActorID   string
	RequestedByActorType domain.ActorType
}

// RecordDTO is a data transfer object representing an audit record for API responses.
// It contains a flattened view of audit data suitable for external consumption.
type RecordDTO struct {
	ID         string `json:"id"`
	ActorID    string `json:"actor_id"`
	ActorType  string `json:"actor_type"`
	Action     string `json:"action"`
	Resource   string `json:"resource"`
	ResourceID string `json:"resource_id"`
	Metadata   string `json:"metadata"`
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
	Status     int    `json:"status"`
	CreatedAt  string `json:"created_at"`
}

// ListRecordsResult contains the paginated result of audit records.
type ListRecordsResult struct {
	Items      []RecordDTO `json:"items"`
	NextCursor string      `json:"next_cursor"`
	HasMore    bool        `json:"has_more"`
	Limit      int         `json:"limit"`
}

// Handle executes the ListRecordsQuery and returns a paginated list of audit records.
// It applies filters based on the query parameters and returns appropriate results
// based on whether the request is scoped to a specific actor.
func (h *ListRecordsHandler) Handle(ctx context.Context, query ListRecordsQuery) (ListRecordsResult, error) {
	filter := domain.RecordFilter{
		ActorID:  query.ActorID,
		Resource: query.Resource,
		Action:   query.Action,
		DateFrom: query.DateFrom,
		DateTo:   query.DateTo,
		Cursor:   query.Cursor,
		Limit:    query.Limit,
	}

	var result pagination.PageResult[*domain.Record]
	var err error

	if query.ActorID != "" && query.RequestedByActorType != domain.ActorUser {
		result, err = h.repo.FindByActorID(ctx, query.ActorID, filter)
	} else {
		result, err = h.repo.FindAll(ctx, filter)
	}

	if err != nil {
		return ListRecordsResult{}, err
	}

	dtos := make([]RecordDTO, len(result.Items))
	for i, record := range result.Items {
		dtos[i] = RecordDTO{
			ID:         record.ID().String(),
			ActorID:    record.ActorID(),
			ActorType:  record.ActorType().String(),
			Action:     record.Action().String(),
			Resource:   record.Resource(),
			ResourceID: record.ResourceID(),
			Metadata:   record.Metadata(),
			IP:         record.IP(),
			UserAgent:  record.UserAgent(),
			Status:     record.Status(),
			CreatedAt:  record.CreatedAt().Format(time.RFC3339),
		}
	}

	return ListRecordsResult{
		Items:      dtos,
		NextCursor: result.NextCursor,
		HasMore:    result.HasMore,
		Limit:      result.Limit,
	}, nil
}

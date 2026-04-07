// Package query provides query handlers for reading task and schedule data.
// This package implements the query side of CQRS for task-related operations,
// handling read-only requests that return task and schedule data transfer objects.
package query

import (
	"context"
	"time"

	"github.com/basilex/skeleton/internal/tasks/domain"
)

// GetTaskQuery represents a query to retrieve a single task by ID.
type GetTaskQuery struct {
	ID domain.TaskID
}

// GetTaskHandler handles queries to retrieve a single task.
type GetTaskHandler struct {
	repo domain.TaskRepository
}

// NewGetTaskHandler creates a new GetTaskHandler with the required repository.
func NewGetTaskHandler(repo domain.TaskRepository) *GetTaskHandler {
	return &GetTaskHandler{repo: repo}
}

// Handle executes the GetTaskQuery and returns the task entity.
func (h *GetTaskHandler) Handle(ctx context.Context, query GetTaskQuery) (*domain.Task, error) {
	return h.repo.GetByID(ctx, query.ID)
}

// ListTasksQuery represents a query to list tasks with optional filtering.
type ListTasksQuery struct {
	TaskType *domain.TaskType
	Status   *domain.TaskStatus
	Priority *domain.TaskPriority
	FromDate *time.Time
	ToDate   *time.Time
	Limit    int
	Cursor   *string
}

// ListTasksHandler handles queries to retrieve a list of tasks.
type ListTasksHandler struct {
	repo domain.TaskRepository
}

// NewListTasksHandler creates a new ListTasksHandler with the required repository.
func NewListTasksHandler(repo domain.TaskRepository) *ListTasksHandler {
	return &ListTasksHandler{repo: repo}
}

// Handle executes the ListTasksQuery and returns matching tasks.
// It filters by status if provided, otherwise by type, or returns pending tasks.
func (h *ListTasksHandler) Handle(ctx context.Context, query ListTasksQuery) ([]*domain.Task, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 100
	}

	if query.Status != nil {
		return h.repo.GetTasksByStatus(ctx, *query.Status, limit)
	}

	if query.TaskType != nil {
		return h.repo.GetTasksByType(ctx, *query.TaskType, limit)
	}

	return h.repo.GetPendingTasks(ctx, limit)
}

// GetActiveTasksQuery represents a query to retrieve all currently active tasks.
type GetActiveTasksQuery struct{}

// GetActiveTasksHandler handles queries to retrieve all active tasks.
type GetActiveTasksHandler struct {
	repo domain.TaskRepository
}

// NewGetActiveTasksHandler creates a new GetActiveTasksHandler with the required repository.
func NewGetActiveTasksHandler(repo domain.TaskRepository) *GetActiveTasksHandler {
	return &GetActiveTasksHandler{repo: repo}
}

// Handle executes the GetActiveTasksQuery and returns all active tasks.
func (h *GetActiveTasksHandler) Handle(ctx context.Context, query GetActiveTasksQuery) ([]*domain.Task, error) {
	return h.repo.GetActiveTasks(ctx)
}

// TaskStats contains aggregated statistics about tasks.
type TaskStats struct {
	TotalTasks     int64
	CompletedTasks int64
	FailedTasks    int64
	RunningTasks   int64
	AvgDurationMs  int64
	ByType         map[domain.TaskType]TypeStats
}

// TypeStats contains statistics for a specific task type.
type TypeStats struct {
	Total   int64
	Success int64
	Failed  int64
}

// GetTaskStatsQuery represents a query to retrieve task statistics within a date range.
type GetTaskStatsQuery struct {
	FromDate time.Time
	ToDate   time.Time
}

// GetTaskStatsHandler handles queries to retrieve aggregated task statistics.
type GetTaskStatsHandler struct {
	repo domain.TaskRepository
}

// NewGetTaskStatsHandler creates a new GetTaskStatsHandler with the required repository.
func NewGetTaskStatsHandler(repo domain.TaskRepository) *GetTaskStatsHandler {
	return &GetTaskStatsHandler{repo: repo}
}

// Handle executes the GetTaskStatsQuery and returns aggregated task statistics.
// It computes totals by querying tasks in various statuses.
func (h *GetTaskStatsHandler) Handle(ctx context.Context, query GetTaskStatsQuery) (*TaskStats, error) {
	completed, err := h.repo.GetTasksByStatus(ctx, domain.TaskStatusCompleted, 1000)
	if err != nil {
		return nil, err
	}

	failed, err := h.repo.GetTasksByStatus(ctx, domain.TaskStatusFailed, 1000)
	if err != nil {
		return nil, err
	}

	running, err := h.repo.GetTasksByStatus(ctx, domain.TaskStatusRunning, 1000)
	if err != nil {
		return nil, err
	}

	stats := &TaskStats{
		TotalTasks:     int64(len(completed) + len(failed) + len(running)),
		CompletedTasks: int64(len(completed)),
		FailedTasks:    int64(len(failed)),
		RunningTasks:   int64(len(running)),
		ByType:         make(map[domain.TaskType]TypeStats),
	}

	return stats, nil
}

// GetScheduleQuery represents a query to retrieve a schedule by ID or name.
type GetScheduleQuery struct {
	ID   *domain.ScheduleID
	Name *string
}

// GetScheduleHandler handles queries to retrieve a single schedule.
type GetScheduleHandler struct {
	repo domain.ScheduleRepository
}

// NewGetScheduleHandler creates a new GetScheduleHandler with the required repository.
func NewGetScheduleHandler(repo domain.ScheduleRepository) *GetScheduleHandler {
	return &GetScheduleHandler{repo: repo}
}

// Handle executes the GetScheduleQuery and returns the schedule.
// It looks up by ID if provided, otherwise by name.
func (h *GetScheduleHandler) Handle(ctx context.Context, query GetScheduleQuery) (*domain.TaskSchedule, error) {
	if query.ID != nil {
		return h.repo.GetByID(ctx, *query.ID)
	}
	if query.Name != nil {
		return h.repo.GetByName(ctx, *query.Name)
	}
	return nil, domain.ErrScheduleNotFound
}

// ListSchedulesQuery represents a query to list all schedules.
type ListSchedulesQuery struct{}

// ListSchedulesHandler handles queries to list all schedules.
type ListSchedulesHandler struct {
	repo domain.ScheduleRepository
}

// NewListSchedulesHandler creates a new ListSchedulesHandler with the required repository.
func NewListSchedulesHandler(repo domain.ScheduleRepository) *ListSchedulesHandler {
	return &ListSchedulesHandler{repo: repo}
}

// Handle executes the ListSchedulesQuery and returns all schedules.
func (h *ListSchedulesHandler) Handle(ctx context.Context, query ListSchedulesQuery) ([]*domain.TaskSchedule, error) {
	return h.repo.List(ctx)
}

// ListDeadLettersQuery represents a query to list dead letter tasks with pagination.
type ListDeadLettersQuery struct {
	Limit  int
	Offset int
}

// ListDeadLettersHandler handles queries to list dead letter tasks.
// Dead letters are tasks that have exceeded their retry limits.
type ListDeadLettersHandler struct {
	repo domain.DeadLetterRepository
}

// NewListDeadLettersHandler creates a new ListDeadLettersHandler with the required repository.
func NewListDeadLettersHandler(repo domain.DeadLetterRepository) *ListDeadLettersHandler {
	return &ListDeadLettersHandler{repo: repo}
}

// Handle executes the ListDeadLettersQuery and returns paginated dead letter tasks.
func (h *ListDeadLettersHandler) Handle(ctx context.Context, query ListDeadLettersQuery) ([]*domain.DeadLetterTask, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 100
	}
	return h.repo.List(ctx, limit, query.Offset)
}

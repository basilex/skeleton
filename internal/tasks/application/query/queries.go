package query

import (
	"context"
	"time"

	"github.com/basilex/skeleton/internal/tasks/domain"
)

type GetTaskQuery struct {
	ID domain.TaskID
}

type GetTaskHandler struct {
	repo domain.TaskRepository
}

func NewGetTaskHandler(repo domain.TaskRepository) *GetTaskHandler {
	return &GetTaskHandler{repo: repo}
}

func (h *GetTaskHandler) Handle(ctx context.Context, query GetTaskQuery) (*domain.Task, error) {
	return h.repo.GetByID(ctx, query.ID)
}

type ListTasksQuery struct {
	TaskType *domain.TaskType
	Status   *domain.TaskStatus
	Priority *domain.TaskPriority
	FromDate *time.Time
	ToDate   *time.Time
	Limit    int
	Cursor   *string
}

type ListTasksHandler struct {
	repo domain.TaskRepository
}

func NewListTasksHandler(repo domain.TaskRepository) *ListTasksHandler {
	return &ListTasksHandler{repo: repo}
}

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

type GetActiveTasksQuery struct{}

type GetActiveTasksHandler struct {
	repo domain.TaskRepository
}

func NewGetActiveTasksHandler(repo domain.TaskRepository) *GetActiveTasksHandler {
	return &GetActiveTasksHandler{repo: repo}
}

func (h *GetActiveTasksHandler) Handle(ctx context.Context, query GetActiveTasksQuery) ([]*domain.Task, error) {
	return h.repo.GetActiveTasks(ctx)
}

type TaskStats struct {
	TotalTasks     int64
	CompletedTasks int64
	FailedTasks    int64
	RunningTasks   int64
	AvgDurationMs  int64
	ByType         map[domain.TaskType]TypeStats
}

type TypeStats struct {
	Total   int64
	Success int64
	Failed  int64
}

type GetTaskStatsQuery struct {
	FromDate time.Time
	ToDate   time.Time
}

type GetTaskStatsHandler struct {
	repo domain.TaskRepository
}

func NewGetTaskStatsHandler(repo domain.TaskRepository) *GetTaskStatsHandler {
	return &GetTaskStatsHandler{repo: repo}
}

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

type GetScheduleQuery struct {
	ID   *domain.ScheduleID
	Name *string
}

type GetScheduleHandler struct {
	repo domain.ScheduleRepository
}

func NewGetScheduleHandler(repo domain.ScheduleRepository) *GetScheduleHandler {
	return &GetScheduleHandler{repo: repo}
}

func (h *GetScheduleHandler) Handle(ctx context.Context, query GetScheduleQuery) (*domain.TaskSchedule, error) {
	if query.ID != nil {
		return h.repo.GetByID(ctx, *query.ID)
	}
	if query.Name != nil {
		return h.repo.GetByName(ctx, *query.Name)
	}
	return nil, domain.ErrScheduleNotFound
}

type ListSchedulesQuery struct{}

type ListSchedulesHandler struct {
	repo domain.ScheduleRepository
}

func NewListSchedulesHandler(repo domain.ScheduleRepository) *ListSchedulesHandler {
	return &ListSchedulesHandler{repo: repo}
}

func (h *ListSchedulesHandler) Handle(ctx context.Context, query ListSchedulesQuery) ([]*domain.TaskSchedule, error) {
	return h.repo.List(ctx)
}

type ListDeadLettersQuery struct {
	Limit  int
	Offset int
}

type ListDeadLettersHandler struct {
	repo domain.DeadLetterRepository
}

func NewListDeadLettersHandler(repo domain.DeadLetterRepository) *ListDeadLettersHandler {
	return &ListDeadLettersHandler{repo: repo}
}

func (h *ListDeadLettersHandler) Handle(ctx context.Context, query ListDeadLettersQuery) ([]*domain.DeadLetterTask, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = 100
	}
	return h.repo.List(ctx, limit, query.Offset)
}

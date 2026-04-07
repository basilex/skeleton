package http

import (
	"time"

	"github.com/basilex/skeleton/internal/tasks/domain"
)

type CreateTaskRequest struct {
	TaskType    string                 `json:"task_type"`
	Payload     map[string]interface{} `json:"payload"`
	Priority    string                 `json:"priority"`
	ScheduledAt *time.Time             `json:"scheduled_at"`
	MaxAttempts *int                   `json:"max_attempts"`
}

type CreateTaskResponse struct {
	TaskID string `json:"task_id"`
}

type TaskResponse struct {
	ID          string                 `json:"id"`
	TaskType    string                 `json:"task_type"`
	Status      string                 `json:"status"`
	Priority    string                 `json:"priority"`
	Payload     map[string]interface{} `json:"payload"`
	Result      *TaskResultResponse    `json:"result,omitempty"`
	Error       *TaskErrorResponse     `json:"error,omitempty"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	ScheduledAt string                 `json:"scheduled_at"`
	StartedAt   *string                `json:"started_at,omitempty"`
	CompletedAt *string                `json:"completed_at,omitempty"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

type TaskResultResponse struct {
	Data       map[string]interface{} `json:"data,omitempty"`
	OutputPath string                 `json:"output_path,omitempty"`
	DurationMs int64                  `json:"duration_ms"`
}

type TaskErrorResponse struct {
	Code    string            `json:"code,omitempty"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type CreateScheduleRequest struct {
	Name     string                 `json:"name"`
	TaskType string                 `json:"task_type"`
	Payload  map[string]interface{} `json:"payload"`
	Cron     string                 `json:"cron"`
	Timezone string                 `json:"timezone"`
}

type CreateScheduleResponse struct {
	ScheduleID string `json:"schedule_id"`
}

type ScheduleResponse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	TaskType  string                 `json:"task_type"`
	Payload   map[string]interface{} `json:"payload"`
	Cron      string                 `json:"cron"`
	Timezone  string                 `json:"timezone"`
	LastRunAt *string                `json:"last_run_at,omitempty"`
	NextRunAt *string                `json:"next_run_at,omitempty"`
	IsActive  bool                   `json:"is_active"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

type DeadLetterResponse struct {
	ID           string       `json:"id"`
	OriginalTask TaskResponse `json:"original_task"`
	FailedAt     string       `json:"failed_at"`
	Reason       string       `json:"reason"`
	Reviewed     bool         `json:"reviewed"`
	ReviewedAt   *string      `json:"reviewed_at,omitempty"`
	ReviewedBy   *string      `json:"reviewed_by,omitempty"`
	Action       string       `json:"action"`
	CreatedAt    string       `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func taskToResponse(task *domain.Task) TaskResponse {
	var startedAt, completedAt *string
	if task.StartedAt() != nil {
		s := task.StartedAt().Format(time.RFC3339)
		startedAt = &s
	}
	if task.CompletedAt() != nil {
		c := task.CompletedAt().Format(time.RFC3339)
		completedAt = &c
	}

	var result *TaskResultResponse
	if task.Result() != nil {
		result = &TaskResultResponse{
			Data:       task.Result().Data,
			OutputPath: task.Result().OutputPath,
			DurationMs: task.Result().DurationMs,
		}
	}

	var taskError *TaskErrorResponse
	if task.Error() != nil {
		taskError = &TaskErrorResponse{
			Code:    task.Error().Code,
			Message: task.Error().Message,
			Details: task.Error().Details,
		}
	}

	return TaskResponse{
		ID:          task.ID().String(),
		TaskType:    task.Type().String(),
		Status:      task.Status().String(),
		Priority:    task.Priority().String(),
		Payload:     task.Payload(),
		Result:      result,
		Error:       taskError,
		Attempts:    task.Attempts(),
		MaxAttempts: task.MaxAttempts(),
		ScheduledAt: task.ScheduledAt().Format(time.RFC3339),
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		CreatedAt:   task.CreatedAt().Format(time.RFC3339),
		UpdatedAt:   task.UpdatedAt().Format(time.RFC3339),
	}
}

func scheduleToResponse(schedule *domain.TaskSchedule) ScheduleResponse {
	var lastRunAt, nextRunAt *string
	if schedule.LastRunAt() != nil {
		l := schedule.LastRunAt().Format(time.RFC3339)
		lastRunAt = &l
	}
	if schedule.NextRunAt() != nil {
		n := schedule.NextRunAt().Format(time.RFC3339)
		nextRunAt = &n
	}

	return ScheduleResponse{
		ID:        schedule.ID().String(),
		Name:      schedule.Name(),
		TaskType:  schedule.TaskType().String(),
		Payload:   schedule.Payload(),
		Cron:      schedule.Cron(),
		Timezone:  schedule.Timezone(),
		LastRunAt: lastRunAt,
		NextRunAt: nextRunAt,
		IsActive:  schedule.IsActive(),
		CreatedAt: schedule.CreatedAt().Format(time.RFC3339),
		UpdatedAt: schedule.UpdatedAt().Format(time.RFC3339),
	}
}

func deadLetterToResponse(dl *domain.DeadLetterTask) DeadLetterResponse {
	var reviewedAt, reviewedBy *string
	if dl.ReviewedAt() != nil {
		r := dl.ReviewedAt().Format(time.RFC3339)
		reviewedAt = &r
	}
	if dl.ReviewedBy() != nil {
		reviewedBy = dl.ReviewedBy()
	}

	return DeadLetterResponse{
		ID:           dl.ID().String(),
		OriginalTask: taskToResponse(dl.OriginalTask()),
		FailedAt:     dl.FailedAt().Format(time.RFC3339),
		Reason:       dl.Reason(),
		Reviewed:     dl.IsReviewed(),
		ReviewedAt:   reviewedAt,
		ReviewedBy:   reviewedBy,
		Action:       dl.Action().String(),
		CreatedAt:    dl.CreatedAt().Format(time.RFC3339),
	}
}

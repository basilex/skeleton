package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/tasks/domain"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	payload, err := json.Marshal(task.Payload())
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var result *domain.TaskResult
	var resultJSON []byte
	if task.Result() != nil {
		result = task.Result()
		resultJSON, err = json.Marshal(result)
		if err != nil {
			return fmt.Errorf("marshal result: %w", err)
		}
	}

	var taskError *domain.TaskError
	var errorJSON []byte
	if task.Error() != nil {
		taskError = task.Error()
		errorJSON, err = json.Marshal(taskError)
		if err != nil {
			return fmt.Errorf("marshal error: %w", err)
		}
	}

	query := `
		INSERT INTO tasks (
			id, type, status, priority, payload, result, error_code, error_message, error_details,
			attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var startedAt, completedAt interface{}
	if task.StartedAt() != nil {
		startedAt = task.StartedAt().Format(time.RFC3339)
	}
	if task.CompletedAt() != nil {
		completedAt = task.CompletedAt().Format(time.RFC3339)
	}

	errorCode := ""
	errorMessage := ""
	errorDetails := errorJSON
	if taskError != nil {
		errorCode = taskError.Code
		errorMessage = taskError.Message
	}

	var resultBytes interface{}
	if resultJSON != nil {
		resultBytes = resultJSON
	}

	var errorBytes interface{}
	if errorDetails != nil {
		errorBytes = errorDetails
	}

	_, err = r.db.ExecContext(ctx, query,
		task.ID().String(),
		task.Type().String(),
		task.Status().String(),
		task.Priority().String(),
		payload,
		resultBytes,
		errorCode,
		errorMessage,
		errorBytes,
		task.Attempts(),
		task.MaxAttempts(),
		task.ScheduledAt().Format(time.RFC3339),
		startedAt,
		completedAt,
		task.CreatedAt().Format(time.RFC3339),
		task.UpdatedAt().Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}

	return nil
}

func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	payload, err := json.Marshal(task.Payload())
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var resultJSON []byte
	if task.Result() != nil {
		resultJSON, err = json.Marshal(task.Result())
		if err != nil {
			return fmt.Errorf("marshal result: %w", err)
		}
	}

	errorCode := ""
	errorMessage := ""
	var errorJSON []byte
	if task.Error() != nil {
		errorCode = task.Error().Code
		errorMessage = task.Error().Message
		errorJSON, err = json.Marshal(task.Error().Details)
		if err != nil {
			return fmt.Errorf("marshal error details: %w", err)
		}
	}

	var startedAt, completedAt interface{}
	if task.StartedAt() != nil {
		startedAt = task.StartedAt().Format(time.RFC3339)
	}
	if task.CompletedAt() != nil {
		completedAt = task.CompletedAt().Format(time.RFC3339)
	}

	var resultBytes interface{}
	if resultJSON != nil {
		resultBytes = resultJSON
	}

	var errorBytes interface{}
	if errorJSON != nil {
		errorBytes = errorJSON
	}

	query := `
		UPDATE tasks SET
			status = ?,
			priority = ?,
			payload = ?,
			result = ?,
			error_code = ?,
			error_message = ?,
			error_details = ?,
			attempts = ?,
			scheduled_at = ?,
			started_at = ?,
			completed_at = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err = r.db.ExecContext(ctx, query,
		task.Status().String(),
		task.Priority().String(),
		payload,
		resultBytes,
		errorCode,
		errorMessage,
		errorBytes,
		task.Attempts(),
		task.ScheduledAt().Format(time.RFC3339),
		startedAt,
		completedAt,
		task.UpdatedAt().Format(time.RFC3339),
		task.ID().String(),
	)

	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id domain.TaskID) (*domain.Task, error) {
	query := `
		SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
			   attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM tasks WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id.String())

	var taskID, taskType, status, priority string
	var payload, resultJSON, errorCode, errorMessage, errorDetails []byte
	var startedAt, completedAt []byte
	var attempts, maxAttempts int
	var scheduledAt, createdAt, updatedAt string

	err := row.Scan(
		&taskID, &taskType, &status, &priority, &payload, &resultJSON, &errorCode, &errorMessage, &errorDetails,
		&attempts, &maxAttempts, &scheduledAt, &startedAt, &completedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrTaskNotFound
		}
		return nil, fmt.Errorf("scan task: %w", err)
	}

	var taskPayload domain.TaskPayload
	if err := json.Unmarshal(payload, &taskPayload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	parsedType, _ := domain.ParseTaskType(taskType)

	task, err := domain.NewTask(parsedType, taskPayload)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (r *TaskRepository) GetPendingTasks(ctx context.Context, limit int) ([]*domain.Task, error) {
	query := `
		SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
			   attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM tasks
		WHERE status = ? AND scheduled_at <= ?
		ORDER BY priority DESC, created_at ASC
		LIMIT ?
	`

	return r.getTasksByQuery(ctx, query, domain.TaskStatusPending, time.Now().Format(time.RFC3339), limit)
}

func (r *TaskRepository) GetTasksByStatus(ctx context.Context, status domain.TaskStatus, limit int) ([]*domain.Task, error) {
	query := `
		SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
			   attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM tasks
		WHERE status = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	return r.getTasksByQuery(ctx, query, status.String(), limit)
}

func (r *TaskRepository) GetTasksByType(ctx context.Context, taskType domain.TaskType, limit int) ([]*domain.Task, error) {
	query := `
		SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
			   attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM tasks
		WHERE type = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	return r.getTasksByQuery(ctx, query, taskType.String(), limit)
}

func (r *TaskRepository) GetScheduledTasks(ctx context.Context, before time.Time, limit int) ([]*domain.Task, error) {
	query := `
		SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
			   attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM tasks
		WHERE status = ? AND scheduled_at <= ?
		ORDER BY scheduled_at ASC
		LIMIT ?
	`

	return r.getTasksByQuery(ctx, query, domain.TaskStatusPending, before.Format(time.RFC3339), limit)
}

func (r *TaskRepository) GetActiveTasks(ctx context.Context) ([]*domain.Task, error) {
	query := `
		SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
			   attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM tasks
		WHERE status = ?
		ORDER BY started_at ASC
	`

	tasks, err := r.getTasksByQuery(ctx, query, domain.TaskStatusRunning, 1000)
	return tasks, err
}

func (r *TaskRepository) GetStalledTasks(ctx context.Context, olderThan time.Duration, limit int) ([]*domain.Task, error) {
	cutoff := time.Now().Add(-olderThan)
	query := `
		SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
			   attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM tasks
		WHERE status = ? AND started_at < ?
		ORDER BY started_at ASC
		LIMIT ?
	`

	return r.getTasksByQuery(ctx, query, domain.TaskStatusRunning, cutoff.Format(time.RFC3339), limit)
}

func (r *TaskRepository) Delete(ctx context.Context, id domain.TaskID) error {
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

func (r *TaskRepository) DeleteCompletedTasks(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	query := `DELETE FROM tasks WHERE status = ? AND completed_at < ?`
	result, err := r.db.ExecContext(ctx, query, domain.TaskStatusCompleted, cutoff.Format(time.RFC3339))
	if err != nil {
		return 0, fmt.Errorf("delete completed tasks: %w", err)
	}
	return result.RowsAffected()
}

func (r *TaskRepository) getTasksByQuery(ctx context.Context, query string, args ...interface{}) ([]*domain.Task, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task, err := r.scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *TaskRepository) scanTask(rows *sql.Rows) (*domain.Task, error) {
	var taskID, taskType, status, priority string
	var payload, resultJSON, errorCode, errorMessage, errorDetails []byte
	var startedAt, completedAt []byte
	var attempts, maxAttempts int
	var scheduledAt, createdAt, updatedAt string

	err := rows.Scan(
		&taskID, &taskType, &status, &priority, &payload, &resultJSON, &errorCode, &errorMessage, &errorDetails,
		&attempts, &maxAttempts, &scheduledAt, &startedAt, &completedAt, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan task: %w", err)
	}

	var taskPayload domain.TaskPayload
	if err := json.Unmarshal(payload, &taskPayload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	parsedType, _ := domain.ParseTaskType(taskType)

	task, err := domain.NewTask(parsedType, taskPayload)
	if err != nil {
		return nil, err
	}

	return task, nil
}

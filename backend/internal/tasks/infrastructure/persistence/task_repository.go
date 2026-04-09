package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/tasks/domain"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	pool *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{
		pool: pool,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

type taskDTO struct {
	ID           string     `db:"id"`
	Type         string     `db:"type"`
	Status       string     `db:"status"`
	Priority     string     `db:"priority"`
	Payload      []byte     `db:"payload"`
	Result       []byte     `db:"result"`
	ErrorCode    string     `db:"error_code"`
	ErrorMessage string     `db:"error_message"`
	ErrorDetails []byte     `db:"error_details"`
	Attempts     int        `db:"attempts"`
	MaxAttempts  int        `db:"max_attempts"`
	ScheduledAt  time.Time  `db:"scheduled_at"`
	StartedAt    *time.Time `db:"started_at"`
	CompletedAt  *time.Time `db:"completed_at"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
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

	var taskError *domain.TaskError
	var errorJSON []byte
	if task.Error() != nil {
		taskError = task.Error()
		errorJSON, err = json.Marshal(taskError.Details)
		if err != nil {
			return fmt.Errorf("marshal error details: %w", err)
		}
	}

	errorCode := ""
	errorMessage := ""
	if taskError != nil {
		errorCode = taskError.Code
		errorMessage = taskError.Message
	}

	query, args, err := r.psql.Insert("tasks").
		Columns("id", "type", "status", "priority", "payload", "result",
			"error_code", "error_message", "error_details",
			"attempts", "max_attempts", "scheduled_at", "started_at",
			"completed_at", "created_at", "updated_at").
		Values(task.ID().String(), task.Type().String(), task.Status().String(),
			task.Priority().String(), payload, resultJSON,
			errorCode, errorMessage, errorJSON,
			task.Attempts(), task.MaxAttempts(), task.ScheduledAt(),
			task.StartedAt(), task.CompletedAt(),
			task.CreatedAt(), task.UpdatedAt()).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
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

	query, args, err := r.psql.Update("tasks").
		Set("status", task.Status().String()).
		Set("priority", task.Priority().String()).
		Set("payload", payload).
		Set("result", resultJSON).
		Set("error_code", errorCode).
		Set("error_message", errorMessage).
		Set("error_details", errorJSON).
		Set("attempts", task.Attempts()).
		Set("scheduled_at", task.ScheduledAt()).
		Set("started_at", task.StartedAt()).
		Set("completed_at", task.CompletedAt()).
		Set("updated_at", task.UpdatedAt()).
		Where(sq.Eq{"id": task.ID().String()}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id domain.TaskID) (*domain.Task, error) {
	var dto taskDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
			   attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM tasks WHERE id = $1`,
		id.String())
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	return r.dtoToDomain(dto)
}

func (r *TaskRepository) GetPendingTasks(ctx context.Context, limit int) ([]*domain.Task, error) {
	var dtos []taskDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
			   attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM tasks
		WHERE status = $1 AND scheduled_at <= $2
		ORDER BY priority DESC, created_at ASC
		LIMIT $3`,
		domain.TaskStatusPending, time.Now(), limit)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}

	return r.dtosToDomains(dtos)
}

func (r *TaskRepository) GetTasksByStatus(ctx context.Context, status domain.TaskStatus, limit int) ([]*domain.Task, error) {
	var dtos []taskDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
			   attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM tasks
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2`,
		status.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}

	return r.dtosToDomains(dtos)
}

func (r *TaskRepository) GetTasksByType(ctx context.Context, taskType domain.TaskType, limit int) ([]*domain.Task, error) {
	var dtos []taskDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
			   attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM tasks
		WHERE type = $1
		ORDER BY created_at DESC
		LIMIT $2`,
		taskType.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}

	return r.dtosToDomains(dtos)
}

func (r *TaskRepository) GetScheduledTasks(ctx context.Context, before time.Time, limit int) ([]*domain.Task, error) {
	var dtos []taskDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
			   attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM tasks
		WHERE status = $1 AND scheduled_at <= $2
		ORDER BY scheduled_at ASC
		LIMIT $3`,
		domain.TaskStatusPending, before, limit)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}

	return r.dtosToDomains(dtos)
}

func (r *TaskRepository) GetActiveTasks(ctx context.Context) ([]*domain.Task, error) {
	var dtos []taskDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
			   attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM tasks
		WHERE status = $1
		ORDER BY started_at ASC`,
		domain.TaskStatusRunning)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}

	return r.dtosToDomains(dtos)
}

func (r *TaskRepository) GetStalledTasks(ctx context.Context, olderThan time.Duration, limit int) ([]*domain.Task, error) {
	cutoff := time.Now().Add(-olderThan)
	var dtos []taskDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, type, status, priority, payload, result, error_code, error_message, error_details,
			   attempts, max_attempts, scheduled_at, started_at, completed_at, created_at, updated_at
		FROM tasks
		WHERE status = $1 AND started_at < $2
		ORDER BY started_at ASC
		LIMIT $3`,
		domain.TaskStatusRunning, cutoff, limit)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}

	return r.dtosToDomains(dtos)
}

func (r *TaskRepository) Delete(ctx context.Context, id domain.TaskID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	return nil
}

func (r *TaskRepository) DeleteCompletedTasks(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE status = $1 AND completed_at < $2`,
		domain.TaskStatusCompleted, cutoff)
	if err != nil {
		return 0, fmt.Errorf("delete completed tasks: %w", err)
	}
	return result.RowsAffected(), nil
}

func (r *TaskRepository) dtoToDomain(dto taskDTO) (*domain.Task, error) {
	var payload domain.TaskPayload
	if err := json.Unmarshal(dto.Payload, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	parsedType, _ := domain.ParseTaskType(dto.Type)

	task, err := domain.NewTask(parsedType, payload)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (r *TaskRepository) dtosToDomains(dtos []taskDTO) ([]*domain.Task, error) {
	tasks := make([]*domain.Task, 0, len(dtos))
	for _, dto := range dtos {
		task, err := r.dtoToDomain(dto)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

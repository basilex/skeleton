package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/tasks/domain"
)

// ScheduleRepository implements the task schedule repository interface
// using SQL database storage.
type ScheduleRepository struct {
	db *sql.DB
}

// NewScheduleRepository creates a new schedule repository with the provided database connection.
func NewScheduleRepository(db *sql.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

// Create persists a new task schedule to the database.
func (r *ScheduleRepository) Create(ctx context.Context, schedule *domain.TaskSchedule) error {
	payload, err := json.Marshal(schedule.Payload())
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var lastRunAt, nextRunAt interface{}
	if schedule.LastRunAt() != nil {
		lastRunAt = schedule.LastRunAt().Format(time.RFC3339)
	}
	if schedule.NextRunAt() != nil {
		nextRunAt = schedule.NextRunAt().Format(time.RFC3339)
	}

	query := `
		INSERT INTO task_schedules (
			id, name, task_type, payload, cron, timezone, last_run_at, next_run_at, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		schedule.ID().String(),
		schedule.Name(),
		schedule.TaskType().String(),
		payload,
		schedule.Cron(),
		schedule.Timezone(),
		lastRunAt,
		nextRunAt,
		schedule.IsActive(),
		schedule.CreatedAt().Format(time.RFC3339),
		schedule.UpdatedAt().Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("insert schedule: %w", err)
	}

	return nil
}

// Update modifies an existing task schedule in the database.
func (r *ScheduleRepository) Update(ctx context.Context, schedule *domain.TaskSchedule) error {
	payload, err := json.Marshal(schedule.Payload())
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	var lastRunAt, nextRunAt interface{}
	if schedule.LastRunAt() != nil {
		lastRunAt = schedule.LastRunAt().Format(time.RFC3339)
	}
	if schedule.NextRunAt() != nil {
		nextRunAt = schedule.NextRunAt().Format(time.RFC3339)
	}

	query := `
		UPDATE task_schedules SET
			name = ?,
			task_type = ?,
			payload = ?,
			cron = ?,
			timezone = ?,
			last_run_at = ?,
			next_run_at = ?,
			is_active = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err = r.db.ExecContext(ctx, query,
		schedule.Name(),
		schedule.TaskType().String(),
		payload,
		schedule.Cron(),
		schedule.Timezone(),
		lastRunAt,
		nextRunAt,
		schedule.IsActive(),
		schedule.UpdatedAt().Format(time.RFC3339),
		schedule.ID().String(),
	)

	if err != nil {
		return fmt.Errorf("update schedule: %w", err)
	}

	return nil
}

// GetByID retrieves a task schedule by its unique identifier.
// Returns domain.ErrScheduleNotFound if no matching schedule exists.
func (r *ScheduleRepository) GetByID(ctx context.Context, id domain.ScheduleID) (*domain.TaskSchedule, error) {
	query := `
		SELECT id, name, task_type, payload, cron, timezone, last_run_at, next_run_at, is_active, created_at, updated_at
		FROM task_schedules WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id.String())

	return r.scanSchedule(row)
}

// GetByName retrieves a task schedule by its unique name.
// Returns domain.ErrScheduleNotFound if no matching schedule exists.
func (r *ScheduleRepository) GetByName(ctx context.Context, name string) (*domain.TaskSchedule, error) {
	query := `
		SELECT id, name, task_type, payload, cron, timezone, last_run_at, next_run_at, is_active, created_at, updated_at
		FROM task_schedules WHERE name = ?
	`

	row := r.db.QueryRowContext(ctx, query, name)

	return r.scanSchedule(row)
}

// GetActiveSchedules retrieves all active schedules ordered by name.
func (r *ScheduleRepository) GetActiveSchedules(ctx context.Context) ([]*domain.TaskSchedule, error) {
	query := `
		SELECT id, name, task_type, payload, cron, timezone, last_run_at, next_run_at, is_active, created_at, updated_at
		FROM task_schedules WHERE is_active = 1
		ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*domain.TaskSchedule
	for rows.Next() {
		schedule, err := r.scanScheduleFromRows(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// List retrieves all schedules ordered by name.
func (r *ScheduleRepository) List(ctx context.Context) ([]*domain.TaskSchedule, error) {
	query := `
		SELECT id, name, task_type, payload, cron, timezone, last_run_at, next_run_at, is_active, created_at, updated_at
		FROM task_schedules ORDER BY name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*domain.TaskSchedule
	for rows.Next() {
		schedule, err := r.scanScheduleFromRows(rows)
		if err != nil {
			return nil, err
		}
		schedules = append(schedules, schedule)
	}

	return schedules, nil
}

// Delete removes a task schedule from the database.
func (r *ScheduleRepository) Delete(ctx context.Context, id domain.ScheduleID) error {
	query := `DELETE FROM task_schedules WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}
	return nil
}

// scanSchedule converts a database row into a domain TaskSchedule entity.
func (r *ScheduleRepository) scanSchedule(row *sql.Row) (*domain.TaskSchedule, error) {
	var id, name, taskType, cron, timezone string
	var payload []byte
	var lastRunAt, nextRunAt []byte
	var isActive bool
	var createdAt, updatedAt string

	err := row.Scan(
		&id, &name, &taskType, &payload, &cron, &timezone, &lastRunAt, &nextRunAt, &isActive, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrScheduleNotFound
		}
		return nil, fmt.Errorf("scan schedule: %w", err)
	}

	var taskPayload domain.TaskPayload
	if err := json.Unmarshal(payload, &taskPayload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	parsedType, _ := domain.ParseTaskType(taskType)

	schedule, err := domain.NewTaskSchedule(name, parsedType, cron, taskPayload)
	if err != nil {
		return nil, err
	}

	if !isActive {
		schedule.Deactivate()
	}

	return schedule, nil
}

// scanScheduleFromRows converts database rows into a domain TaskSchedule entity.
func (r *ScheduleRepository) scanScheduleFromRows(rows *sql.Rows) (*domain.TaskSchedule, error) {
	var id, name, taskType, cron, timezone string
	var payload []byte
	var lastRunAt, nextRunAt []byte
	var isActive bool
	var createdAt, updatedAt string

	err := rows.Scan(
		&id, &name, &taskType, &payload, &cron, &timezone, &lastRunAt, &nextRunAt, &isActive, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan schedule: %w", err)
	}

	var taskPayload domain.TaskPayload
	if err := json.Unmarshal(payload, &taskPayload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	parsedType, _ := domain.ParseTaskType(taskType)

	schedule, err := domain.NewTaskSchedule(name, parsedType, cron, taskPayload)
	if err != nil {
		return nil, err
	}

	if !isActive {
		schedule.Deactivate()
	}

	return schedule, nil
}

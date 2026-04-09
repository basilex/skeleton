package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/tasks/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleRepository struct {
	pool *pgxpool.Pool
}

func NewScheduleRepository(pool *pgxpool.Pool) *ScheduleRepository {
	return &ScheduleRepository{pool: pool}
}

func (r *ScheduleRepository) Create(ctx context.Context, schedule *domain.TaskSchedule) error {
	payload, err := json.Marshal(schedule.Payload())
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	query := `
		INSERT INTO task_schedules (
			id, name, task_type, payload, cron, timezone, last_run_at, next_run_at, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	var lastRunAt, nextRunAt *time.Time
	if schedule.LastRunAt() != nil {
		lastRunAt = schedule.LastRunAt()
	}
	if schedule.NextRunAt() != nil {
		nextRunAt = schedule.NextRunAt()
	}

	_, err = r.pool.Exec(ctx, query,
		schedule.ID().String(),
		schedule.Name(),
		schedule.TaskType().String(),
		payload,
		schedule.Cron(),
		schedule.Timezone(),
		lastRunAt,
		nextRunAt,
		schedule.IsActive(),
		schedule.CreatedAt(),
		schedule.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("insert schedule: %w", err)
	}

	return nil
}

func (r *ScheduleRepository) Update(ctx context.Context, schedule *domain.TaskSchedule) error {
	payload, err := json.Marshal(schedule.Payload())
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	query := `
		UPDATE task_schedules SET
			name = $1,
			task_type = $2,
			payload = $3,
			cron = $4,
			timezone = $5,
			last_run_at = $6,
			next_run_at = $7,
			is_active = $8,
			updated_at = $9
		WHERE id = $10
	`

	var lastRunAt, nextRunAt *time.Time
	if schedule.LastRunAt() != nil {
		lastRunAt = schedule.LastRunAt()
	}
	if schedule.NextRunAt() != nil {
		nextRunAt = schedule.NextRunAt()
	}

	_, err = r.pool.Exec(ctx, query,
		schedule.Name(),
		schedule.TaskType().String(),
		payload,
		schedule.Cron(),
		schedule.Timezone(),
		lastRunAt,
		nextRunAt,
		schedule.IsActive(),
		schedule.UpdatedAt(),
		schedule.ID().String(),
	)

	if err != nil {
		return fmt.Errorf("update schedule: %w", err)
	}

	return nil
}

func (r *ScheduleRepository) GetByID(ctx context.Context, id domain.ScheduleID) (*domain.TaskSchedule, error) {
	query := `
		SELECT id, name, task_type, payload, cron, timezone, last_run_at, next_run_at, is_active, created_at, updated_at
		FROM task_schedules WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id.String())

	var idStr, name, taskType, cron, timezone string
	var payload []byte
	var lastRunAt, nextRunAt *time.Time
	var isActive bool
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&idStr, &name, &taskType, &payload, &cron, &timezone, &lastRunAt, &nextRunAt, &isActive, &createdAt, &updatedAt,
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

func (r *ScheduleRepository) GetByName(ctx context.Context, name string) (*domain.TaskSchedule, error) {
	query := `
		SELECT id, name, task_type, payload, cron, timezone, last_run_at, next_run_at, is_active, created_at, updated_at
		FROM task_schedules WHERE name = $1
	`

	row := r.pool.QueryRow(ctx, query, name)

	var idStr, taskType, cron, timezone string
	var payload []byte
	var lastRunAt, nextRunAt *time.Time
	var isActive bool
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&idStr, &name, &taskType, &payload, &cron, &timezone, &lastRunAt, &nextRunAt, &isActive, &createdAt, &updatedAt,
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

func (r *ScheduleRepository) GetActiveSchedules(ctx context.Context) ([]*domain.TaskSchedule, error) {
	query := `
		SELECT id, name, task_type, payload, cron, timezone, last_run_at, next_run_at, is_active, created_at, updated_at
		FROM task_schedules WHERE is_active = true
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*domain.TaskSchedule
	for rows.Next() {
		var idStr, name, taskType, cron, timezone string
		var payload []byte
		var lastRunAt, nextRunAt *time.Time
		var isActive bool
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&idStr, &name, &taskType, &payload, &cron, &timezone, &lastRunAt, &nextRunAt, &isActive, &createdAt, &updatedAt,
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

		schedules = append(schedules, schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return schedules, nil
}

func (r *ScheduleRepository) List(ctx context.Context) ([]*domain.TaskSchedule, error) {
	query := `
		SELECT id, name, task_type, payload, cron, timezone, last_run_at, next_run_at, is_active, created_at, updated_at
		FROM task_schedules ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*domain.TaskSchedule
	for rows.Next() {
		var idStr, name, taskType, cron, timezone string
		var payload []byte
		var lastRunAt, nextRunAt *time.Time
		var isActive bool
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&idStr, &name, &taskType, &payload, &cron, &timezone, &lastRunAt, &nextRunAt, &isActive, &createdAt, &updatedAt,
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

		schedules = append(schedules, schedule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return schedules, nil
}

func (r *ScheduleRepository) Delete(ctx context.Context, id domain.ScheduleID) error {
	query := `DELETE FROM task_schedules WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}
	return nil
}

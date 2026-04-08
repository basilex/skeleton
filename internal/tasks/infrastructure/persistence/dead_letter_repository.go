package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/tasks/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeadLetterRepository struct {
	pool *pgxpool.Pool
}

func NewDeadLetterRepository(pool *pgxpool.Pool) *DeadLetterRepository {
	return &DeadLetterRepository{pool: pool}
}

func (r *DeadLetterRepository) Create(ctx context.Context, deadLetter *domain.DeadLetterTask) error {
	originalTaskJSON, err := json.Marshal(deadLetter.OriginalTask())
	if err != nil {
		return fmt.Errorf("marshal original task: %w", err)
	}

	query := `
		INSERT INTO dead_letters (
			id, original_task_id, original_task, failed_at, reason, reviewed, reviewed_at, reviewed_by, action, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	var reviewedAt *time.Time
	var reviewedBy *string
	if deadLetter.ReviewedAt() != nil {
		reviewedAt = deadLetter.ReviewedAt()
	}
	if deadLetter.ReviewedBy() != nil {
		reviewedBy = deadLetter.ReviewedBy()
	}

	_, err = r.pool.Exec(ctx, query,
		deadLetter.ID().String(),
		deadLetter.OriginalTask().ID().String(),
		originalTaskJSON,
		deadLetter.FailedAt(),
		deadLetter.Reason(),
		deadLetter.IsReviewed(),
		reviewedAt,
		reviewedBy,
		deadLetter.Action().String(),
		deadLetter.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("insert dead letter: %w", err)
	}

	return nil
}

func (r *DeadLetterRepository) GetByID(ctx context.Context, id domain.DeadLetterID) (*domain.DeadLetterTask, error) {
	query := `
		SELECT id, original_task_id, original_task, failed_at, reason, reviewed, reviewed_at, reviewed_by, action, created_at
		FROM dead_letters WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id.String())

	var idStr, originalTaskID, reason, action string
	var originalTaskJSON []byte
	var failedAt, createdAt time.Time
	var reviewed bool
	var reviewedAt *time.Time
	var reviewedBy *string

	err := row.Scan(
		&idStr, &originalTaskID, &originalTaskJSON, &failedAt, &reason, &reviewed, &reviewedAt, &reviewedBy, &action, &createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan dead letter: %w", err)
	}

	var task domain.Task
	if err := json.Unmarshal(originalTaskJSON, &task); err != nil {
		return nil, fmt.Errorf("unmarshal original task: %w", err)
	}

	dl := domain.NewDeadLetterTask(&task, reason)

	return dl, nil
}

func (r *DeadLetterRepository) List(ctx context.Context, limit int, offset int) ([]*domain.DeadLetterTask, error) {
	query := `
		SELECT id, original_task_id, original_task, failed_at, reason, reviewed, reviewed_at, reviewed_by, action, created_at
		FROM dead_letters
		ORDER BY failed_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query dead letters: %w", err)
	}
	defer rows.Close()

	var deadLetters []*domain.DeadLetterTask
	for rows.Next() {
		var idStr, originalTaskID, reason, action string
		var originalTaskJSON []byte
		var failedAt, createdAt time.Time
		var reviewed bool
		var reviewedAt *time.Time
		var reviewedBy *string

		err := rows.Scan(
			&idStr, &originalTaskID, &originalTaskJSON, &failedAt, &reason, &reviewed, &reviewedAt, &reviewedBy, &action, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan dead letter: %w", err)
		}

		var task domain.Task
		if err := json.Unmarshal(originalTaskJSON, &task); err != nil {
			return nil, fmt.Errorf("unmarshal original task: %w", err)
		}

		dl := domain.NewDeadLetterTask(&task, reason)

		deadLetters = append(deadLetters, dl)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return deadLetters, nil
}

func (r *DeadLetterRepository) MarkReviewed(ctx context.Context, id domain.DeadLetterID, action domain.DeadLetterAction, reviewedBy *string) error {
	query := `
		UPDATE dead_letters SET reviewed = $1, reviewed_at = $2, reviewed_by = $3, action = $4 WHERE id = $5
	`

	reviewedAt := time.Now()

	_, err := r.pool.Exec(ctx, query, true, reviewedAt, reviewedBy, action.String(), id.String())
	if err != nil {
		return fmt.Errorf("update dead letter: %w", err)
	}

	return nil
}

func (r *DeadLetterRepository) Delete(ctx context.Context, id domain.DeadLetterID) error {
	query := `DELETE FROM dead_letters WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("delete dead letter: %w", err)
	}
	return nil
}

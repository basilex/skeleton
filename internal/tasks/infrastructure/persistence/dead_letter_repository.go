package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/tasks/domain"
)

type DeadLetterRepository struct {
	db *sql.DB
}

func NewDeadLetterRepository(db *sql.DB) *DeadLetterRepository {
	return &DeadLetterRepository{db: db}
}

func (r *DeadLetterRepository) Create(ctx context.Context, deadLetter *domain.DeadLetterTask) error {
	originalTaskJSON, err := json.Marshal(deadLetter.OriginalTask())
	if err != nil {
		return fmt.Errorf("marshal original task: %w", err)
	}

	query := `
		INSERT INTO dead_letters (
			id, original_task_id, original_task, failed_at, reason, reviewed, reviewed_at, reviewed_by, action, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var reviewedAt, reviewedBy interface{}
	if deadLetter.ReviewedAt() != nil {
		reviewedAt = deadLetter.ReviewedAt().Format(time.RFC3339)
	}
	if deadLetter.ReviewedBy() != nil {
		reviewedBy = *deadLetter.ReviewedBy()
	}

	_, err = r.db.ExecContext(ctx, query,
		deadLetter.ID().String(),
		deadLetter.OriginalTask().ID().String(),
		originalTaskJSON,
		deadLetter.FailedAt().Format(time.RFC3339),
		deadLetter.Reason(),
		deadLetter.IsReviewed(),
		reviewedAt,
		reviewedBy,
		deadLetter.Action().String(),
		deadLetter.CreatedAt().Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("insert dead letter: %w", err)
	}

	return nil
}

func (r *DeadLetterRepository) GetByID(ctx context.Context, id domain.DeadLetterID) (*domain.DeadLetterTask, error) {
	query := `
		SELECT id, original_task_id, original_task, failed_at, reason, reviewed, reviewed_at, reviewed_by, action, created_at
		FROM dead_letters WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id.String())

	return r.scanDeadLetter(row)
}

func (r *DeadLetterRepository) List(ctx context.Context, limit int, offset int) ([]*domain.DeadLetterTask, error) {
	query := `
		SELECT id, original_task_id, original_task, failed_at, reason, reviewed, reviewed_at, reviewed_by, action, created_at
		FROM dead_letters
		ORDER BY failed_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query dead letters: %w", err)
	}
	defer rows.Close()

	var deadLetters []*domain.DeadLetterTask
	for rows.Next() {
		dl, err := r.scanDeadLetterFromRows(rows)
		if err != nil {
			return nil, err
		}
		deadLetters = append(deadLetters, dl)
	}

	return deadLetters, nil
}

func (r *DeadLetterRepository) MarkReviewed(ctx context.Context, id domain.DeadLetterID, action domain.DeadLetterAction, reviewedBy *string) error {
	query := `
		UPDATE dead_letters SET
			reviewed = ?,
			reviewed_at = ?,
			reviewed_by = ?,
			action = ?
		WHERE id = ?
	`

	reviewedAt := time.Now().Format(time.RFC3339)

	_, err := r.db.ExecContext(ctx, query, true, reviewedAt, reviewedBy, action.String(), id.String())
	if err != nil {
		return fmt.Errorf("update dead letter: %w", err)
	}

	return nil
}

func (r *DeadLetterRepository) Delete(ctx context.Context, id domain.DeadLetterID) error {
	query := `DELETE FROM dead_letters WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("delete dead letter: %w", err)
	}
	return nil
}

func (r *DeadLetterRepository) scanDeadLetter(row *sql.Row) (*domain.DeadLetterTask, error) {
	var id, originalTaskID, reason, action string
	var originalTaskJSON []byte
	var failedAtStr, createdAtStr string
	var reviewed bool
	var reviewedAtStr, reviewedBy sql.NullString

	err := row.Scan(
		&id, &originalTaskID, &originalTaskJSON, &failedAtStr, &reason, &reviewed, &reviewedAtStr, &reviewedBy, &action, &createdAtStr,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrDeadLetterNotFound
		}
		return nil, fmt.Errorf("scan dead letter: %w", err)
	}

	var task domain.Task
	if err := json.Unmarshal(originalTaskJSON, &task); err != nil {
		return nil, fmt.Errorf("unmarshal original task: %w", err)
	}

	dl := domain.NewDeadLetterTask(&task, reason)

	return dl, nil
}

func (r *DeadLetterRepository) scanDeadLetterFromRows(rows *sql.Rows) (*domain.DeadLetterTask, error) {
	var id, originalTaskID, reason, action string
	var originalTaskJSON []byte
	var failedAtStr, createdAtStr string
	var reviewed bool
	var reviewedAtStr, reviewedBy sql.NullString

	err := rows.Scan(
		&id, &originalTaskID, &originalTaskJSON, &failedAtStr, &reviewed, &reviewedAtStr, &reviewedBy, &action, &createdAtStr,
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

package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/files/domain"
)

// ProcessingRepository implements domain.ProcessingRepository using SQLite.
type ProcessingRepository struct {
	db *sql.DB
}

// NewProcessingRepository creates a new processing repository.
func NewProcessingRepository(db *sql.DB) *ProcessingRepository {
	return &ProcessingRepository{db: db}
}

// Create inserts a new processing record.
func (r *ProcessingRepository) Create(ctx context.Context, processing *domain.FileProcessing) error {
	query := `
		INSERT INTO file_processings (
			id, file_id, operation, options, status, result_file_id, error,
			started_at, completed_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	optionsJSON, err := json.Marshal(processing.Options())
	if err != nil {
		return fmt.Errorf("marshal options: %w", err)
	}

	var resultFileID, errorMsg, startedAt, completedAt *string
	if processing.ResultFileID() != nil {
		rfid := processing.ResultFileID().String()
		resultFileID = &rfid
	}
	if processing.Error() != nil {
		errorMsg = processing.Error()
	}
	if processing.StartedAt() != nil {
		sa := processing.StartedAt().Format(time.RFC3339)
		startedAt = &sa
	}
	if processing.CompletedAt() != nil {
		ca := processing.CompletedAt().Format(time.RFC3339)
		completedAt = &ca
	}

	_, err = r.db.ExecContext(ctx, query,
		processing.ID().String(),
		processing.FileID().String(),
		string(processing.Operation()),
		optionsJSON,
		string(processing.Status()),
		resultFileID,
		errorMsg,
		startedAt,
		completedAt,
		processing.CreatedAt().Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("insert processing: %w", err)
	}

	return nil
}

// Update updates an existing processing record.
func (r *ProcessingRepository) Update(ctx context.Context, processing *domain.FileProcessing) error {
	query := `
		UPDATE file_processings SET
			status = ?, result_file_id = ?, error = ?,
			started_at = ?, completed_at = ?
		WHERE id = ?
	`

	var resultFileID, errorMsg, startedAt, completedAt *string
	if processing.ResultFileID() != nil {
		rfid := processing.ResultFileID().String()
		resultFileID = &rfid
	}
	if processing.Error() != nil {
		errorMsg = processing.Error()
	}
	if processing.StartedAt() != nil {
		sa := processing.StartedAt().Format(time.RFC3339)
		startedAt = &sa
	}
	if processing.CompletedAt() != nil {
		ca := processing.CompletedAt().Format(time.RFC3339)
		completedAt = &ca
	}

	result, err := r.db.ExecContext(ctx, query,
		string(processing.Status()),
		resultFileID,
		errorMsg,
		startedAt,
		completedAt,
		processing.ID().String(),
	)

	if err != nil {
		return fmt.Errorf("update processing: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrProcessingNotFound
	}

	return nil
}

// GetByID retrieves a processing record by ID.
func (r *ProcessingRepository) GetByID(ctx context.Context, id domain.ProcessingID) (*domain.FileProcessing, error) {
	query := `
		SELECT id, file_id, operation, options, status, result_file_id, error,
			started_at, completed_at, created_at
		FROM file_processings WHERE id = ?
	`

	return r.scanProcessing(ctx, query, id.String())
}

// GetByFileID retrieves all processing records for a file.
func (r *ProcessingRepository) GetByFileID(ctx context.Context, fileID domain.FileID) ([]*domain.FileProcessing, error) {
	query := `
		SELECT id, file_id, operation, options, status, result_file_id, error,
			started_at, completed_at, created_at
		FROM file_processings WHERE file_id = ? ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, fileID.String())
	if err != nil {
		return nil, fmt.Errorf("query processings: %w", err)
	}
	defer rows.Close()

	return r.scanProcessings(rows)
}

// GetPending retrieves all pending processing tasks.
func (r *ProcessingRepository) GetPending(ctx context.Context, limit int) ([]*domain.FileProcessing, error) {
	query := `
		SELECT id, file_id, operation, options, status, result_file_id, error,
			started_at, completed_at, created_at
		FROM file_processings WHERE status = ? ORDER BY created_at ASC LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, string(domain.ProcessingPending), limit)
	if err != nil {
		return nil, fmt.Errorf("query pending: %w", err)
	}
	defer rows.Close()

	return r.scanProcessings(rows)
}

// GetByStatus retrieves processing records by status.
func (r *ProcessingRepository) GetByStatus(ctx context.Context, status domain.ProcessingStatus, limit, offset int) ([]*domain.FileProcessing, error) {
	query := `
		SELECT id, file_id, operation, options, status, result_file_id, error,
			started_at, completed_at, created_at
		FROM file_processings WHERE status = ? ORDER BY created_at DESC LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, string(status), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query by status: %w", err)
	}
	defer rows.Close()

	return r.scanProcessings(rows)
}

// Delete deletes a processing record by ID.
func (r *ProcessingRepository) Delete(ctx context.Context, id domain.ProcessingID) error {
	query := `DELETE FROM file_processings WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("delete processing: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrProcessingNotFound
	}

	return nil
}

// DeleteByFileID deletes all processing records for a file.
func (r *ProcessingRepository) DeleteByFileID(ctx context.Context, fileID domain.FileID) error {
	query := `DELETE FROM file_processings WHERE file_id = ?`

	_, err := r.db.ExecContext(ctx, query, fileID.String())
	if err != nil {
		return fmt.Errorf("delete processings by file: %w", err)
	}

	return nil
}

// Count returns the total number of processing records matching the filter.
func (r *ProcessingRepository) Count(ctx context.Context, filter *domain.ProcessingFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM file_processings WHERE 1=1`
	args := []interface{}{}

	if filter != nil {
		if filter.FileID != nil {
			query += ` AND file_id = ?`
			args = append(args, filter.FileID.String())
		}
		if filter.Status != nil {
			query += ` AND status = ?`
			args = append(args, string(*filter.Status))
		}
		if filter.Operation != nil {
			query += ` AND operation = ?`
			args = append(args, string(*filter.Operation))
		}
	}

	var count int64
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count processings: %w", err)
	}

	return count, nil
}

// List retrieves processing records matching the filter with pagination.
func (r *ProcessingRepository) List(ctx context.Context, filter *domain.ProcessingFilter, limit, offset int) ([]*domain.FileProcessing, error) {
	query := `
		SELECT id, file_id, operation, options, status, result_file_id, error,
			started_at, completed_at, created_at
		FROM file_processings WHERE 1=1
	`
	args := []interface{}{}

	if filter != nil {
		if filter.FileID != nil {
			query += ` AND file_id = ?`
			args = append(args, filter.FileID.String())
		}
		if filter.Status != nil {
			query += ` AND status = ?`
			args = append(args, string(*filter.Status))
		}
		if filter.Operation != nil {
			query += ` AND operation = ?`
			args = append(args, string(*filter.Operation))
		}
	}

	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query processings: %w", err)
	}
	defer rows.Close()

	return r.scanProcessings(rows)
}

// scanProcessing scans a single processing record.
func (r *ProcessingRepository) scanProcessing(ctx context.Context, query string, args ...interface{}) (*domain.FileProcessing, error) {
	var processingID, fileID, operation, status string
	var optionsJSON []byte
	var resultFileID, errorMsg, startedAtStr, completedAtStr *string
	var createdAtStr string

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&processingID, &fileID, &operation, &optionsJSON, &status,
		&resultFileID, &errorMsg, &startedAtStr, &completedAtStr, &createdAtStr,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrProcessingNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan processing: %w", err)
	}

	return r.reconstituteProcessing(
		processingID, fileID, operation, optionsJSON, status,
		resultFileID, errorMsg, startedAtStr, completedAtStr, createdAtStr,
	)
}

// scanProcessings scans multiple processing records.
func (r *ProcessingRepository) scanProcessings(rows *sql.Rows) ([]*domain.FileProcessing, error) {
	var processings []*domain.FileProcessing

	for rows.Next() {
		var processingID, fileID, operation, status string
		var optionsJSON []byte
		var resultFileID, errorMsg, startedAtStr, completedAtStr *string
		var createdAtStr string

		if err := rows.Scan(
			&processingID, &fileID, &operation, &optionsJSON, &status,
			&resultFileID, &errorMsg, &startedAtStr, &completedAtStr, &createdAtStr,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		processing, err := r.reconstituteProcessing(
			processingID, fileID, operation, optionsJSON, status,
			resultFileID, errorMsg, startedAtStr, completedAtStr, createdAtStr,
		)
		if err != nil {
			return nil, err
		}

		processings = append(processings, processing)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return processings, nil
}

// reconstituteProcessing reconstructs a FileProcessing from database fields.
func (r *ProcessingRepository) reconstituteProcessing(
	processingID, fileID, operation string, optionsJSON []byte, status string,
	resultFileID, errorMsg *string, startedAtStr, completedAtStr *string, createdAtStr string,
) (*domain.FileProcessing, error) {
	var options domain.ProcessingOptions
	if len(optionsJSON) > 0 {
		if err := json.Unmarshal(optionsJSON, &options); err != nil {
			return nil, fmt.Errorf("unmarshal options: %w", err)
		}
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	var startedAt, completedAt *time.Time
	if startedAtStr != nil && *startedAtStr != "" {
		t, err := time.Parse(time.RFC3339, *startedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse started_at: %w", err)
		}
		startedAt = &t
	}

	if completedAtStr != nil && *completedAtStr != "" {
		t, err := time.Parse(time.RFC3339, *completedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse completed_at: %w", err)
		}
		completedAt = &t
	}

	var rfID *domain.FileID
	if resultFileID != nil && *resultFileID != "" {
		fid := domain.FileID(*resultFileID)
		rfID = &fid
	}

	return domain.ReconstituteFileProcessing(
		domain.ProcessingID(processingID),
		domain.FileID(fileID),
		domain.ProcessingOperation(operation),
		options,
		domain.ProcessingStatus(status),
		rfID,
		errorMsg,
		startedAt,
		completedAt,
		createdAt,
	), nil
}

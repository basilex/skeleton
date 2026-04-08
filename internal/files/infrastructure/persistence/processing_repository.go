package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/files/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProcessingRepository struct {
	pool *pgxpool.Pool
}

func NewProcessingRepository(pool *pgxpool.Pool) *ProcessingRepository {
	return &ProcessingRepository{pool: pool}
}

func (r *ProcessingRepository) Create(ctx context.Context, processing *domain.FileProcessing) error {
	query := `
		INSERT INTO file_processings (
			id, file_id, operation, options, status, result_file_id, error,
			started_at, completed_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	optionsJSON, err := json.Marshal(processing.Options())
	if err != nil {
		return fmt.Errorf("marshal options: %w", err)
	}

	var resultFileID, errorMsg *string
	var startedAt, completedAt *time.Time
	if processing.ResultFileID() != nil {
		rfid := processing.ResultFileID().String()
		resultFileID = &rfid
	}
	if processing.Error() != nil {
		errorMsg = processing.Error()
	}
	if processing.StartedAt() != nil {
		startedAt = processing.StartedAt()
	}
	if processing.CompletedAt() != nil {
		completedAt = processing.CompletedAt()
	}

	_, err = r.pool.Exec(ctx, query,
		processing.ID().String(),
		processing.FileID().String(),
		string(processing.Operation()),
		optionsJSON,
		string(processing.Status()),
		resultFileID,
		errorMsg,
		startedAt,
		completedAt,
		processing.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("insert processing: %w", err)
	}

	return nil
}

func (r *ProcessingRepository) Update(ctx context.Context, processing *domain.FileProcessing) error {
	query := `
		UPDATE file_processings SET
			status = $1, result_file_id = $2, error = $3, started_at = $4, completed_at = $5
		WHERE id = $6
	`

	var resultFileID, errorMsg *string
	var startedAt, completedAt *time.Time
	if processing.ResultFileID() != nil {
		rfid := processing.ResultFileID().String()
		resultFileID = &rfid
	}
	if processing.Error() != nil {
		errorMsg = processing.Error()
	}
	if processing.StartedAt() != nil {
		startedAt = processing.StartedAt()
	}
	if processing.CompletedAt() != nil {
		completedAt = processing.CompletedAt()
	}

	result, err := r.pool.Exec(ctx, query,
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

	if result.RowsAffected() == 0 {
		return domain.ErrProcessingNotFound
	}

	return nil
}

func (r *ProcessingRepository) GetByID(ctx context.Context, id domain.ProcessingID) (*domain.FileProcessing, error) {
	query := `
		SELECT id, file_id, operation, options, status, result_file_id, error,
			started_at, completed_at, created_at
		FROM file_processings WHERE id = $1
	`

	return r.scanProcessing(ctx, query, id.String())
}

func (r *ProcessingRepository) GetByFileID(ctx context.Context, fileID domain.FileID) ([]*domain.FileProcessing, error) {
	query := `
		SELECT id, file_id, operation, options, status, result_file_id, error,
			started_at, completed_at, created_at
		FROM file_processings WHERE file_id = $1 ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, fileID.String())
	if err != nil {
		return nil, fmt.Errorf("query processings: %w", err)
	}
	defer rows.Close()

	return r.scanProcessings(rows)
}

func (r *ProcessingRepository) GetPending(ctx context.Context, limit int) ([]*domain.FileProcessing, error) {
	query := `
		SELECT id, file_id, operation, options, status, result_file_id, error,
			started_at, completed_at, created_at
		FROM file_processings WHERE status = $1 ORDER BY created_at ASC LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, string(domain.ProcessingPending), limit)
	if err != nil {
		return nil, fmt.Errorf("query pending: %w", err)
	}
	defer rows.Close()

	return r.scanProcessings(rows)
}

func (r *ProcessingRepository) GetByStatus(ctx context.Context, status domain.ProcessingStatus, limit, offset int) ([]*domain.FileProcessing, error) {
	query := `
		SELECT id, file_id, operation, options, status, result_file_id, error,
			started_at, completed_at, created_at
		FROM file_processings WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, string(status), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query by status: %w", err)
	}
	defer rows.Close()

	return r.scanProcessings(rows)
}

func (r *ProcessingRepository) Delete(ctx context.Context, id domain.ProcessingID) error {
	query := `DELETE FROM file_processings WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("delete processing: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrProcessingNotFound
	}

	return nil
}

func (r *ProcessingRepository) DeleteByFileID(ctx context.Context, fileID domain.FileID) error {
	query := `DELETE FROM file_processings WHERE file_id = $1`

	_, err := r.pool.Exec(ctx, query, fileID.String())
	if err != nil {
		return fmt.Errorf("delete processings by file: %w", err)
	}

	return nil
}

func (r *ProcessingRepository) Count(ctx context.Context, filter *domain.ProcessingFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM file_processings WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if filter != nil {
		if filter.FileID != nil {
			query += fmt.Sprintf(` AND file_id = $%d`, argNum)
			args = append(args, filter.FileID.String())
			argNum++
		}
		if filter.Status != nil {
			query += fmt.Sprintf(` AND status = $%d`, argNum)
			args = append(args, string(*filter.Status))
			argNum++
		}
		if filter.Operation != nil {
			query += fmt.Sprintf(` AND operation = $%d`, argNum)
			args = append(args, string(*filter.Operation))
			argNum++
		}
	}

	var count int64
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count processings: %w", err)
	}

	return count, nil
}

func (r *ProcessingRepository) List(ctx context.Context, filter *domain.ProcessingFilter, limit, offset int) ([]*domain.FileProcessing, error) {
	query := `
		SELECT id, file_id, operation, options, status, result_file_id, error,
			started_at, completed_at, created_at
		FROM file_processings WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if filter != nil {
		if filter.FileID != nil {
			query += fmt.Sprintf(` AND file_id = $%d`, argNum)
			args = append(args, filter.FileID.String())
			argNum++
		}
		if filter.Status != nil {
			query += fmt.Sprintf(` AND status = $%d`, argNum)
			args = append(args, string(*filter.Status))
			argNum++
		}
		if filter.Operation != nil {
			query += fmt.Sprintf(` AND operation = $%d`, argNum)
			args = append(args, string(*filter.Operation))
			argNum++
		}
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query processings: %w", err)
	}
	defer rows.Close()

	return r.scanProcessings(rows)
}

func (r *ProcessingRepository) scanProcessing(ctx context.Context, query string, args ...interface{}) (*domain.FileProcessing, error) {
	var processingID, fileID, operation, status string
	var optionsJSON []byte
	var resultFileID, errorMsg *string
	var startedAt, completedAt *time.Time
	var createdAt time.Time

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&processingID, &fileID, &operation, &optionsJSON, &status,
		&resultFileID, &errorMsg, &startedAt, &completedAt, &createdAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrProcessingNotFound
		}
		return nil, fmt.Errorf("scan processing: %w", err)
	}

	return r.reconstituteProcessing(
		processingID, fileID, operation, optionsJSON, status,
		resultFileID, errorMsg, startedAt, completedAt, createdAt,
	)
}

func (r *ProcessingRepository) scanProcessings(rows pgx.Rows) ([]*domain.FileProcessing, error) {
	var processings []*domain.FileProcessing

	for rows.Next() {
		var processingID, fileID, operation, status string
		var optionsJSON []byte
		var resultFileID, errorMsg *string
		var startedAt, completedAt *time.Time
		var createdAt time.Time

		if err := rows.Scan(
			&processingID, &fileID, &operation, &optionsJSON, &status,
			&resultFileID, &errorMsg, &startedAt, &completedAt, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		processing, err := r.reconstituteProcessing(
			processingID, fileID, operation, optionsJSON, status,
			resultFileID, errorMsg, startedAt, completedAt, createdAt,
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

func (r *ProcessingRepository) reconstituteProcessing(
	processingID, fileID, operation string, optionsJSON []byte, status string,
	resultFileID, errorMsg *string, startedAt, completedAt *time.Time, createdAt time.Time,
) (*domain.FileProcessing, error) {
	var options domain.ProcessingOptions
	if len(optionsJSON) > 0 {
		if err := json.Unmarshal(optionsJSON, &options); err != nil {
			return nil, fmt.Errorf("unmarshal options: %w", err)
		}
	}

	var rfID *domain.FileID
	if resultFileID != nil && *resultFileID != "" {
		fid, parseErr := domain.ParseFileID(*resultFileID)
		if parseErr != nil {
			return nil, fmt.Errorf("parse result file id: %w", parseErr)
		}
		rfID = &fid
	}

	processingIDParsed, err := domain.ParseProcessingID(processingID)
	if err != nil {
		return nil, fmt.Errorf("parse processing id: %w", err)
	}
	fileIDParsed, err := domain.ParseFileID(fileID)
	if err != nil {
		return nil, fmt.Errorf("parse file id: %w", err)
	}

	return domain.ReconstituteFileProcessing(
		processingIDParsed,
		fileIDParsed,
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

package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/files/domain"
	identitydomain "github.com/basilex/skeleton/internal/identity/domain"
)

// FileRepository implements domain.FileRepository using SQLite.
type FileRepository struct {
	db *sql.DB
}

// NewFileRepository creates a new file repository.
func NewFileRepository(db *sql.DB) *FileRepository {
	return &FileRepository{db: db}
}

// Create inserts a new file record.
func (r *FileRepository) Create(ctx context.Context, file *domain.File) error {
	query := `
		INSERT INTO files (
			id, owner_id, filename, stored_name, mime_type, size, path,
			storage_provider, checksum, metadata, access_level,
			uploaded_at, expires_at, processed_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	metadataJSON, err := json.Marshal(file.Metadata())
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	var ownerID *string
	if file.OwnerID() != nil {
		oid := string(*file.OwnerID())
		ownerID = &oid
	}

	var expiresAt, processedAt *string
	if file.ExpiresAt() != nil {
		e := file.ExpiresAt().Format(time.RFC3339)
		expiresAt = &e
	}
	if file.ProcessedAt() != nil {
		p := file.ProcessedAt().Format(time.RFC3339)
		processedAt = &p
	}

	_, err = r.db.ExecContext(ctx, query,
		file.ID().String(),
		ownerID,
		file.Filename(),
		file.StoredName(),
		file.MimeType(),
		file.Size(),
		file.Path(),
		string(file.StorageProvider()),
		file.Checksum(),
		metadataJSON,
		string(file.AccessLevel()),
		file.UploadedAt().Format(time.RFC3339),
		expiresAt,
		processedAt,
		file.CreatedAt().Format(time.RFC3339),
		file.UpdatedAt().Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("insert file: %w", err)
	}

	return nil
}

// GetByID retrieves a file by ID.
func (r *FileRepository) GetByID(ctx context.Context, id domain.FileID) (*domain.File, error) {
	query := `
		SELECT id, owner_id, filename, stored_name, mime_type, size, path,
			storage_provider, checksum, metadata, access_level,
			uploaded_at, expires_at, processed_at, created_at, updated_at
		FROM files WHERE id = ?
	`

	var fileID, filename, storedName, mimeType, path, storageProvider, checksum, accessLevel string
	var ownerID sql.NullString
	var size int64
	var metadataJSON []byte
	var uploadedAt, createdAt, updatedAt string
	var expiresAt, processedAt *string

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&fileID, &ownerID, &filename, &storedName, &mimeType, &size, &path,
		&storageProvider, &checksum, &metadataJSON, &accessLevel,
		&uploadedAt, &expiresAt, &processedAt, &createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrFileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}

	var ownerIDStr string
	if ownerID.Valid {
		ownerIDStr = ownerID.String
	}

	return r.reconstituteFile(
		fileID, ownerIDStr, filename, storedName, mimeType, size, path,
		storageProvider, checksum, metadataJSON, accessLevel,
		uploadedAt, expiresAt, processedAt, createdAt, updatedAt,
	)
}

// Update updates an existing file record.
func (r *FileRepository) Update(ctx context.Context, file *domain.File) error {
	query := `
		UPDATE files SET
			filename = ?, stored_name = ?, mime_type = ?, size = ?, path = ?,
			storage_provider = ?, checksum = ?, metadata = ?, access_level = ?,
			expires_at = ?, processed_at = ?, updated_at = ?
		WHERE id = ?
	`

	metadataJSON, err := json.Marshal(file.Metadata())
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	var expiresAt, processedAt *string
	if file.ExpiresAt() != nil {
		e := file.ExpiresAt().Format(time.RFC3339)
		expiresAt = &e
	}
	if file.ProcessedAt() != nil {
		p := file.ProcessedAt().Format(time.RFC3339)
		processedAt = &p
	}

	result, err := r.db.ExecContext(ctx, query,
		file.Filename(),
		file.StoredName(),
		file.MimeType(),
		file.Size(),
		file.Path(),
		string(file.StorageProvider()),
		file.Checksum(),
		metadataJSON,
		string(file.AccessLevel()),
		expiresAt,
		processedAt,
		time.Now().Format(time.RFC3339),
		file.ID().String(),
	)

	if err != nil {
		return fmt.Errorf("update file: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrFileNotFound
	}

	return nil
}

// Delete deletes a file by ID.
func (r *FileRepository) Delete(ctx context.Context, id domain.FileID) error {
	query := `DELETE FROM files WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrFileNotFound
	}

	return nil
}

// GetByPath retrieves a file by storage path.
func (r *FileRepository) GetByPath(ctx context.Context, path string) (*domain.File, error) {
	query := `
		SELECT id, owner_id, filename, stored_name, mime_type, size, path,
			storage_provider, checksum, metadata, access_level,
			uploaded_at, expires_at, processed_at, created_at, updated_at
		FROM files WHERE path = ?
	`

	var fileID, filename, storedName, mimeType, filePath, storageProvider, checksum, accessLevel string
	var ownerID sql.NullString
	var size int64
	var metadataJSON []byte
	var uploadedAt, createdAt, updatedAt string
	var expiresAt, processedAt *string

	err := r.db.QueryRowContext(ctx, query, path).Scan(
		&fileID, &ownerID, &filename, &storedName, &mimeType, &size, &filePath,
		&storageProvider, &checksum, &metadataJSON, &accessLevel,
		&uploadedAt, &expiresAt, &processedAt, &createdAt, &updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrFileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}

	var ownerIDStr string
	if ownerID.Valid {
		ownerIDStr = ownerID.String
	}

	return r.reconstituteFile(
		fileID, ownerIDStr, filename, storedName, mimeType, size, filePath,
		storageProvider, checksum, metadataJSON, accessLevel,
		uploadedAt, expiresAt, processedAt, createdAt, updatedAt,
	)
}

// GetByOwner retrieves files by owner ID with pagination.
func (r *FileRepository) GetByOwner(ctx context.Context, ownerID string, limit, offset int) ([]*domain.File, error) {
	query := `
		SELECT id, owner_id, filename, stored_name, mime_type, size, path,
			storage_provider, checksum, metadata, access_level,
			uploaded_at, expires_at, processed_at, created_at, updated_at
		FROM files 
		WHERE owner_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query files: %w", err)
	}
	defer rows.Close()

	return r.scanFiles(rows)
}

// GetExpired retrieves expired files.
func (r *FileRepository) GetExpired(ctx context.Context, before time.Time, limit int) ([]*domain.File, error) {
	query := `
		SELECT id, owner_id, filename, stored_name, mime_type, size, path,
			storage_provider, checksum, metadata, access_level,
			uploaded_at, expires_at, processed_at, created_at, updated_at
		FROM files 
		WHERE expires_at IS NOT NULL AND expires_at < ?
		ORDER BY expires_at ASC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, before.Format(time.RFC3339), limit)
	if err != nil {
		return nil, fmt.Errorf("query expired files: %w", err)
	}
	defer rows.Close()

	return r.scanFiles(rows)
}

// List retrieves files matching the filter with pagination.
func (r *FileRepository) List(ctx context.Context, filter *domain.FileFilter, limit, offset int) ([]*domain.File, error) {
	query := `SELECT id, owner_id, filename, stored_name, mime_type, size, path,
		storage_provider, checksum, metadata, access_level,
		uploaded_at, expires_at, processed_at, created_at, updated_at
		FROM files WHERE 1=1`
	args := []interface{}{}

	if filter != nil {
		if filter.OwnerID != nil {
			query += ` AND owner_id = ?`
			args = append(args, *filter.OwnerID)
		}
		if filter.MimeType != nil {
			query += ` AND mime_type LIKE ?`
			args = append(args, *filter.MimeType+`%`)
		}
		if filter.AccessLevel != nil {
			query += ` AND access_level = ?`
			args = append(args, string(*filter.AccessLevel))
		}
		if filter.UploadedFrom != nil {
			query += ` AND uploaded_at >= ?`
			args = append(args, filter.UploadedFrom.Format(time.RFC3339))
		}
		if filter.UploadedTo != nil {
			query += ` AND uploaded_at <= ?`
			args = append(args, filter.UploadedTo.Format(time.RFC3339))
		}
	}

	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query files: %w", err)
	}
	defer rows.Close()

	return r.scanFiles(rows)
}

// DeleteBatch deletes multiple files by IDs.
func (r *FileRepository) DeleteBatch(ctx context.Context, ids []domain.FileID) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `DELETE FROM files WHERE id = ?`)
	if err != nil {
		return fmt.Errorf("prepare stmt: %w", err)
	}
	defer stmt.Close()

	for _, id := range ids {
		if _, err := stmt.ExecContext(ctx, id.String()); err != nil {
			return fmt.Errorf("delete file %s: %w", id, err)
		}
	}

	return tx.Commit()
}

// Count returns the total number of files matching the filter.
func (r *FileRepository) Count(ctx context.Context, filter *domain.FileFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM files WHERE 1=1`
	args := []interface{}{}

	if filter != nil {
		if filter.OwnerID != nil {
			query += ` AND owner_id = ?`
			args = append(args, *filter.OwnerID)
		}
		if filter.MimeType != nil {
			query += ` AND mime_type LIKE ?`
			args = append(args, *filter.MimeType+`%`)
		}
		if filter.AccessLevel != nil {
			query += ` AND access_level = ?`
			args = append(args, string(*filter.AccessLevel))
		}
	}

	var count int64
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count files: %w", err)
	}

	return count, nil
}

// scanFiles scans multiple files from rows.
func (r *FileRepository) scanFiles(rows *sql.Rows) ([]*domain.File, error) {
	var files []*domain.File

	for rows.Next() {
		var fileID, filename, storedName, mimeType, path, storageProvider, checksum, accessLevel string
		var ownerID sql.NullString
		var size int64
		var metadataJSON []byte
		var uploadedAt, createdAt, updatedAt string
		var expiresAt, processedAt *string

		if err := rows.Scan(
			&fileID, &ownerID, &filename, &storedName, &mimeType, &size, &path,
			&storageProvider, &checksum, &metadataJSON, &accessLevel,
			&uploadedAt, &expiresAt, &processedAt, &createdAt, &updatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		var ownerIDStr string
		if ownerID.Valid {
			ownerIDStr = ownerID.String
		}

		file, err := r.reconstituteFile(
			fileID, ownerIDStr, filename, storedName, mimeType, size, path,
			storageProvider, checksum, metadataJSON, accessLevel,
			uploadedAt, expiresAt, processedAt, createdAt, updatedAt,
		)
		if err != nil {
			return nil, err
		}

		files = append(files, file)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return files, nil
}

// reconstituteFile reconstructs a File from database fields.
func (r *FileRepository) reconstituteFile(
	fileID, ownerID, filename, storedName, mimeType string, size int64, path string,
	storageProvider, checksum string, metadataJSON []byte, accessLevel string,
	uploadedAtStr string, expiresAtStr, processedAtStr *string,
	createdAtStr, updatedAtStr string,
) (*domain.File, error) {
	var metadata domain.FileMetadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	uploadedAt, err := time.Parse(time.RFC3339, uploadedAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse uploaded_at: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}

	var expiresAt, processedAt *time.Time
	if expiresAtStr != nil && *expiresAtStr != "" {
		t, err := time.Parse(time.RFC3339, *expiresAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse expires_at: %w", err)
		}
		expiresAt = &t
	}

	if processedAtStr != nil && *processedAtStr != "" {
		t, err := time.Parse(time.RFC3339, *processedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse processed_at: %w", err)
		}
		processedAt = &t
	}

	var ownerIDPtr *identitydomain.UserID
	if ownerID != "" {
		uid := identitydomain.UserID(ownerID)
		ownerIDPtr = &uid
	}

	return domain.ReconstituteFile(
		domain.FileID(fileID),
		ownerIDPtr,
		filename,
		storedName,
		mimeType,
		size,
		path,
		domain.StorageProvider(storageProvider),
		checksum,
		metadata,
		domain.AccessLevel(accessLevel),
		uploadedAt,
		expiresAt,
		processedAt,
		createdAt,
		updatedAt,
	), nil
}

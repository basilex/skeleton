package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/files/domain"
)

// UploadRepository implements domain.UploadRepository using SQLite.
type UploadRepository struct {
	db *sql.DB
}

// NewUploadRepository creates a new upload repository.
func NewUploadRepository(db *sql.DB) *UploadRepository {
	return &UploadRepository{db: db}
}

// Create inserts a new upload record.
func (r *UploadRepository) Create(ctx context.Context, upload *domain.FileUpload) error {
	query := `
		INSERT INTO file_uploads (id, file_id, upload_url, fields, status, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	fieldsJSON, err := json.Marshal(upload.Fields())
	if err != nil {
		return fmt.Errorf("marshal fields: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		upload.ID().String(),
		upload.File().ID().String(),
		upload.UploadURL(),
		fieldsJSON,
		string(upload.Status()),
		upload.ExpiresAt().Format(time.RFC3339),
		upload.CreatedAt().Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("insert upload: %w", err)
	}

	return nil
}

// GetByID retrieves an upload by ID.
func (r *UploadRepository) GetByID(ctx context.Context, id domain.UploadID) (*domain.FileUpload, error) {
	query := `
		SELECT u.id, u.file_id, u.upload_url, u.fields, u.status, u.expires_at, u.created_at,
		       f.owner_id, f.filename, f.stored_name, f.mime_type, f.size, f.path,
		       f.storage_provider, f.checksum, f.access_level, f.uploaded_at, f.created_at
		FROM file_uploads u
		JOIN files f ON u.file_id = f.id
		WHERE u.id = ?
	`

	var uploadID, fileID, uploadURL, status string
	var fieldsJSON []byte
	var expiresAt, createdAt string

	// File fields
	var filename, storedName, mimeType, path, storageProvider, checksum, accessLevel string
	var ownerID sql.NullString
	var size int64
	var uploadedAt, fileCreatedAt string

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&uploadID, &fileID, &uploadURL, &fieldsJSON, &status, &expiresAt, &createdAt,
		&ownerID, &filename, &storedName, &mimeType, &size, &path,
		&storageProvider, &checksum, &accessLevel, &uploadedAt, &fileCreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrUploadNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan upload: %w", err)
	}

	return r.reconstituteUpload(uploadID, fileID, uploadURL, fieldsJSON, status, expiresAt, createdAt,
		ownerID, filename, storedName, mimeType, size, path, storageProvider, checksum, accessLevel, uploadedAt, fileCreatedAt)
}

// GetByFileID retrieves an upload for a specific file.
func (r *UploadRepository) GetByFileID(ctx context.Context, fileID domain.FileID) (*domain.FileUpload, error) {
	query := `
		SELECT u.id, u.file_id, u.upload_url, u.fields, u.status, u.expires_at, u.created_at,
		       f.owner_id, f.filename, f.stored_name, f.mime_type, f.size, f.path,
		       f.storage_provider, f.checksum, f.access_level, f.uploaded_at, f.created_at
		FROM file_uploads u
		JOIN files f ON u.file_id = f.id
		WHERE u.file_id = ? ORDER BY u.created_at DESC LIMIT 1
	`

	var uploadID, fileIDStr, uploadURL, status string
	var fieldsJSON []byte
	var expiresAt, createdAt string

	// File fields
	var filename, storedName, mimeType, path, storageProvider, checksum, accessLevel string
	var ownerID sql.NullString
	var size int64
	var uploadedAt, fileCreatedAt string

	err := r.db.QueryRowContext(ctx, query, fileID.String()).Scan(
		&uploadID, &fileIDStr, &uploadURL, &fieldsJSON, &status, &expiresAt, &createdAt,
		&ownerID, &filename, &storedName, &mimeType, &size, &path,
		&storageProvider, &checksum, &accessLevel, &uploadedAt, &fileCreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrUploadNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan upload: %w", err)
	}

	return r.reconstituteUpload(uploadID, fileIDStr, uploadURL, fieldsJSON, status, expiresAt, createdAt,
		ownerID, filename, storedName, mimeType, size, path, storageProvider, checksum, accessLevel, uploadedAt, fileCreatedAt)
}

// UpdateStatus updates the upload status.
func (r *UploadRepository) UpdateStatus(ctx context.Context, id domain.UploadID, status domain.UploadStatus) error {
	query := `UPDATE file_uploads SET status = ? WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, string(status), id.String())
	if err != nil {
		return fmt.Errorf("update upload status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrUploadNotFound
	}

	return nil
}

// GetExpired retrieves expired uploads.
func (r *UploadRepository) GetExpired(ctx context.Context, before time.Time, limit int) ([]*domain.FileUpload, error) {
	query := `
		SELECT u.id, u.file_id, u.upload_url, u.fields, u.status, u.expires_at, u.created_at,
		       f.owner_id, f.filename, f.stored_name, f.mime_type, f.size, f.path,
		       f.storage_provider, f.checksum, f.access_level, f.uploaded_at, f.created_at
		FROM file_uploads u
		JOIN files f ON u.file_id = f.id
		WHERE u.expires_at < ? ORDER BY u.expires_at ASC LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, before.Format(time.RFC3339), limit)
	if err != nil {
		return nil, fmt.Errorf("query expired uploads: %w", err)
	}
	defer rows.Close()

	var uploads []*domain.FileUpload

	for rows.Next() {
		var uploadID, fileID, uploadURL, status string
		var fieldsJSON []byte
		var expiresAt, createdAt string

		// File fields
		var filename, storedName, mimeType, path, storageProvider, checksum, accessLevel string
		var ownerID sql.NullString
		var size int64
		var uploadedAt, fileCreatedAt string

		if err := rows.Scan(
			&uploadID, &fileID, &uploadURL, &fieldsJSON, &status, &expiresAt, &createdAt,
			&ownerID, &filename, &storedName, &mimeType, &size, &path,
			&storageProvider, &checksum, &accessLevel, &uploadedAt, &fileCreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		upload, err := r.reconstituteUpload(uploadID, fileID, uploadURL, fieldsJSON, status, expiresAt, createdAt,
			ownerID, filename, storedName, mimeType, size, path, storageProvider, checksum, accessLevel, uploadedAt, fileCreatedAt)
		if err != nil {
			return nil, err
		}

		uploads = append(uploads, upload)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return uploads, nil
}

// Delete deletes an upload by ID.
func (r *UploadRepository) Delete(ctx context.Context, id domain.UploadID) error {
	query := `DELETE FROM file_uploads WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("delete upload: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return domain.ErrUploadNotFound
	}

	return nil
}

// DeleteByFileID deletes an upload by file ID.
func (r *UploadRepository) DeleteByFileID(ctx context.Context, fileID domain.FileID) error {
	query := `DELETE FROM file_uploads WHERE file_id = ?`

	_, err := r.db.ExecContext(ctx, query, fileID.String())
	if err != nil {
		return fmt.Errorf("delete upload by file: %w", err)
	}

	return nil
}

// reconstituteUpload reconstructs a FileUpload from database fields.
func (r *UploadRepository) reconstituteUpload(
	uploadID, fileID, uploadURL string, fieldsJSON []byte, status string,
	expiresAtStr, createdAtStr string,
	ownerID sql.NullString, filename, storedName, mimeType string, size int64, path string,
	storageProvider, checksum, accessLevel string, uploadedAtStr, fileCreatedAtStr string,
) (*domain.FileUpload, error) {
	var fields map[string]string
	if len(fieldsJSON) > 0 {
		if err := json.Unmarshal(fieldsJSON, &fields); err != nil {
			return nil, fmt.Errorf("unmarshal fields: %w", err)
		}
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse expires_at: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	uploadedAt, err := time.Parse(time.RFC3339, uploadedAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse uploaded_at: %w", err)
	}

	fileCreatedAt, err := time.Parse(time.RFC3339, fileCreatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse file_created_at: %w", err)
	}

	// Reconstitute file
	file := domain.ReconstituteFile(
		domain.FileID(fileID),
		nil, // owner ID will be set separately if needed
		filename,
		storedName,
		mimeType,
		size,
		path,
		domain.StorageProvider(storageProvider),
		checksum,
		domain.FileMetadata{},
		domain.AccessLevel(accessLevel),
		uploadedAt,
		nil, // expires at
		nil, // processed at
		fileCreatedAt,
		fileCreatedAt,
	)

	upload := domain.ReconstituteFileUpload(
		domain.UploadID(uploadID),
		file,
		uploadURL,
		fields,
		domain.UploadStatus(status),
		expiresAt,
		createdAt,
	)

	return upload, nil
}

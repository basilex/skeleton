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

type UploadRepository struct {
	pool *pgxpool.Pool
}

func NewUploadRepository(pool *pgxpool.Pool) *UploadRepository {
	return &UploadRepository{pool: pool}
}

func (r *UploadRepository) Create(ctx context.Context, upload *domain.FileUpload) error {
	query := `
		INSERT INTO file_uploads (id, file_id, upload_url, fields, status, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	fieldsJSON, err := json.Marshal(upload.Fields())
	if err != nil {
		return fmt.Errorf("marshal fields: %w", err)
	}

	_, err = r.pool.Exec(ctx, query,
		upload.ID().String(),
		upload.File().ID().String(),
		upload.UploadURL(),
		fieldsJSON,
		string(upload.Status()),
		upload.ExpiresAt(),
		upload.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("insert upload: %w", err)
	}

	return nil
}

func (r *UploadRepository) GetByID(ctx context.Context, id domain.UploadID) (*domain.FileUpload, error) {
	query := `
		SELECT u.id, u.file_id, u.upload_url, u.fields, u.status, u.expires_at, u.created_at,
		       f.owner_id, f.filename, f.stored_name, f.mime_type, f.size, f.path,
		       f.storage_provider, f.checksum, f.access_level, f.uploaded_at, f.created_at
		FROM file_uploads u
		JOIN files f ON u.file_id = f.id
		WHERE u.id = $1
	`

	var uploadID, fileID, uploadURL, status string
	var fieldsJSON []byte
	var expiresAt, createdAt time.Time

	var filename, storedName, mimeType, path, storageProvider, checksum, accessLevel string
	var ownerID *string
	var size int64
	var uploadedAt, fileCreatedAt time.Time

	err := r.pool.QueryRow(ctx, query, id.String()).Scan(
		&uploadID, &fileID, &uploadURL, &fieldsJSON, &status, &expiresAt, &createdAt,
		&ownerID, &filename, &storedName, &mimeType, &size, &path,
		&storageProvider, &checksum, &accessLevel, &uploadedAt, &fileCreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUploadNotFound
		}
		return nil, fmt.Errorf("scan upload: %w", err)
	}

	return r.reconstituteUpload(uploadID, fileID, uploadURL, fieldsJSON, status, expiresAt, createdAt,
		ownerID, filename, storedName, mimeType, size, path, storageProvider, checksum, accessLevel, uploadedAt, fileCreatedAt)
}

func (r *UploadRepository) GetByFileID(ctx context.Context, fileID domain.FileID) (*domain.FileUpload, error) {
	query := `
		SELECT u.id, u.file_id, u.upload_url, u.fields, u.status, u.expires_at, u.created_at,
		       f.owner_id, f.filename, f.stored_name, f.mime_type, f.size, f.path,
		       f.storage_provider, f.checksum, f.access_level, f.uploaded_at, f.created_at
		FROM file_uploads u
		JOIN files f ON u.file_id = f.id
		WHERE u.file_id = $1 ORDER BY u.created_at DESC LIMIT 1
	`

	var uploadID, fileIDStr, uploadURL, status string
	var fieldsJSON []byte
	var expiresAt, createdAt time.Time

	var filename, storedName, mimeType, path, storageProvider, checksum, accessLevel string
	var ownerID *string
	var size int64
	var uploadedAt, fileCreatedAt time.Time

	err := r.pool.QueryRow(ctx, query, fileID.String()).Scan(
		&uploadID, &fileIDStr, &uploadURL, &fieldsJSON, &status, &expiresAt, &createdAt,
		&ownerID, &filename, &storedName, &mimeType, &size, &path,
		&storageProvider, &checksum, &accessLevel, &uploadedAt, &fileCreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrUploadNotFound
		}
		return nil, fmt.Errorf("scan upload: %w", err)
	}

	return r.reconstituteUpload(uploadID, fileIDStr, uploadURL, fieldsJSON, status, expiresAt, createdAt,
		ownerID, filename, storedName, mimeType, size, path, storageProvider, checksum, accessLevel, uploadedAt, fileCreatedAt)
}

func (r *UploadRepository) UpdateStatus(ctx context.Context, id domain.UploadID, status domain.UploadStatus) error {
	query := `UPDATE file_uploads SET status = $1 WHERE id = $2`

	result, err := r.pool.Exec(ctx, query, string(status), id.String())
	if err != nil {
		return fmt.Errorf("update upload status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUploadNotFound
	}

	return nil
}

func (r *UploadRepository) GetExpired(ctx context.Context, before time.Time, limit int) ([]*domain.FileUpload, error) {
	query := `
		SELECT u.id, u.file_id, u.upload_url, u.fields, u.status, u.expires_at, u.created_at,
		       f.owner_id, f.filename, f.stored_name, f.mime_type, f.size, f.path,
		       f.storage_provider, f.checksum, f.access_level, f.uploaded_at, f.created_at
		FROM file_uploads u
		JOIN files f ON u.file_id = f.id
		WHERE u.expires_at < $1 ORDER BY u.expires_at ASC LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, before, limit)
	if err != nil {
		return nil, fmt.Errorf("query expired uploads: %w", err)
	}
	defer rows.Close()

	var uploads []*domain.FileUpload

	for rows.Next() {
		var uploadID, fileID, uploadURL, status string
		var fieldsJSON []byte
		var expiresAt, createdAt time.Time

		var filename, storedName, mimeType, path, storageProvider, checksum, accessLevel string
		var ownerID *string
		var size int64
		var uploadedAt, fileCreatedAt time.Time

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

func (r *UploadRepository) Delete(ctx context.Context, id domain.UploadID) error {
	query := `DELETE FROM file_uploads WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("delete upload: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrUploadNotFound
	}

	return nil
}

func (r *UploadRepository) DeleteByFileID(ctx context.Context, fileID domain.FileID) error {
	query := `DELETE FROM file_uploads WHERE file_id = $1`

	_, err := r.pool.Exec(ctx, query, fileID.String())
	if err != nil {
		return fmt.Errorf("delete upload by file: %w", err)
	}

	return nil
}

func (r *UploadRepository) reconstituteUpload(
	uploadID, fileID, uploadURL string, fieldsJSON []byte, status string,
	expiresAt, createdAt time.Time,
	ownerID *string, filename, storedName, mimeType string, size int64, path string,
	storageProvider, checksum, accessLevel string, uploadedAt, fileCreatedAt time.Time,
) (*domain.FileUpload, error) {
	var fields map[string]string
	if len(fieldsJSON) > 0 {
		if err := json.Unmarshal(fieldsJSON, &fields); err != nil {
			return nil, fmt.Errorf("unmarshal fields: %w", err)
		}
	}

	fileIDParsed, err := domain.ParseFileID(fileID)
	if err != nil {
		return nil, fmt.Errorf("parse file id: %w", err)
	}

	file := domain.ReconstituteFile(
		fileIDParsed,
		nil,
		filename,
		storedName,
		mimeType,
		size,
		path,
		domain.StorageProvider(storageProvider),
		checksum,
		domain.FileMetadata{},
		domain.AccessLevel(accessLevel),
		domain.ScanStatusPending,
		"",
		nil,
		uploadedAt,
		nil,
		nil,
		fileCreatedAt,
		fileCreatedAt,
	)

	uploadIDParsed, err := domain.ParseUploadID(uploadID)
	if err != nil {
		return nil, fmt.Errorf("parse upload id: %w", err)
	}

	upload := domain.ReconstituteFileUpload(
		uploadIDParsed,
		file,
		uploadURL,
		fields,
		domain.UploadStatus(status),
		expiresAt,
		createdAt,
	)

	return upload, nil
}

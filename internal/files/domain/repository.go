package domain

import (
	"context"
	"time"
)

// FileRepository defines the interface for file persistence operations.
type FileRepository interface {
	// Create creates a new file record.
	Create(ctx context.Context, file *File) error

	// Update updates an existing file record.
	Update(ctx context.Context, file *File) error

	// GetByID retrieves a file by its ID.
	GetByID(ctx context.Context, id FileID) (*File, error)

	// GetByPath retrieves a file by its storage path.
	GetByPath(ctx context.Context, path string) (*File, error)

	// GetByOwner retrieves all files owned by a specific user.
	// Use limit and offset for pagination.
	GetByOwner(ctx context.Context, ownerID string, limit, offset int) ([]*File, error)

	// GetExpired retrieves all files that have expired before the given time.
	GetExpired(ctx context.Context, before time.Time, limit int) ([]*File, error)

	// Delete deletes a file record by ID.
	Delete(ctx context.Context, id FileID) error

	// DeleteBatch deletes multiple file records by IDs.
	DeleteBatch(ctx context.Context, ids []FileID) error

	// Count returns the total number of files matching the filter.
	Count(ctx context.Context, filter *FileFilter) (int64, error)

	// List retrieves files matching the filter with pagination.
	List(ctx context.Context, filter *FileFilter, limit, offset int) ([]*File, error)
}

// FileFilter represents filter options for file queries.
type FileFilter struct {
	OwnerID      *string      // Filter by owner ID
	MimeType     *string      // Filter by MIME type prefix
	AccessLevel  *AccessLevel // Filter by access level
	UploadedFrom *time.Time   // Filter by upload date (from)
	UploadedTo   *time.Time   // Filter by upload date (to)
	ExpiresFrom  *time.Time   // Filter by expiration date (from)
	ExpiresTo    *time.Time   // Filter by expiration date (to)
}

// UploadRepository defines the interface for upload URL persistence operations.
type UploadRepository interface {
	// Create creates a new upload record.
	Create(ctx context.Context, upload *FileUpload) error

	// GetByID retrieves an upload by its ID.
	GetByID(ctx context.Context, id UploadID) (*FileUpload, error)

	// GetByFileID retrieves the upload for a specific file.
	GetByFileID(ctx context.Context, fileID FileID) (*FileUpload, error)

	// UpdateStatus updates the upload status.
	UpdateStatus(ctx context.Context, id UploadID, status UploadStatus) error

	// GetExpired retrieves all uploads that have expired before the given time.
	GetExpired(ctx context.Context, before time.Time, limit int) ([]*FileUpload, error)

	// Delete deletes an upload record by ID.
	Delete(ctx context.Context, id UploadID) error

	// DeleteByFileID deletes an upload record by file ID.
	DeleteByFileID(ctx context.Context, fileID FileID) error
}

// ProcessingRepository defines the interface for file processing persistence operations.
type ProcessingRepository interface {
	// Create creates a new processing record.
	Create(ctx context.Context, processing *FileProcessing) error

	// Update updates an existing processing record.
	Update(ctx context.Context, processing *FileProcessing) error

	// GetByID retrieves a processing record by its ID.
	GetByID(ctx context.Context, id ProcessingID) (*FileProcessing, error)

	// GetByFileID retrieves all processing records for a specific file.
	GetByFileID(ctx context.Context, fileID FileID) ([]*FileProcessing, error)

	// GetPending retrieves all pending processing tasks.
	// Use limit to control batch size.
	GetPending(ctx context.Context, limit int) ([]*FileProcessing, error)

	// GetByStatus retrieves processing records by status.
	GetByStatus(ctx context.Context, status ProcessingStatus, limit, offset int) ([]*FileProcessing, error)

	// Delete deletes a processing record by ID.
	Delete(ctx context.Context, id ProcessingID) error

	// DeleteByFileID deletes all processing records for a specific file.
	DeleteByFileID(ctx context.Context, fileID FileID) error

	// Count returns the total number of processing records matching the filter.
	Count(ctx context.Context, filter *ProcessingFilter) (int64, error)

	// List retrieves processing records matching the filter with pagination.
	List(ctx context.Context, filter *ProcessingFilter, limit, offset int) ([]*FileProcessing, error)
}

// ProcessingFilter represents filter options for processing queries.
type ProcessingFilter struct {
	FileID      *FileID              // Filter by file ID
	Status      *ProcessingStatus    // Filter by status
	Operation   *ProcessingOperation // Filter by operation type
	CreatedFrom *time.Time           // Filter by creation date (from)
	CreatedTo   *time.Time           // Filter by creation date (to)
}

package query

import (
	"context"
	"time"

	"github.com/basilex/skeleton/internal/files/domain"
)

// GetFileQuery retrieves file metadata by ID.
type GetFileQuery struct {
	FileID string
}

// GetFileHandler handles get file query.
type GetFileHandler func(ctx context.Context, query GetFileQuery) (*FileDTO, error)

// ListFilesQuery lists files with pagination and filters.
type ListFilesQuery struct {
	OwnerID      *string
	MimeType     *string
	AccessLevel  *domain.AccessLevel
	UploadedFrom *time.Time
	UploadedTo   *time.Time
	Limit        int
	Cursor       *string
}

// ListFilesResult contains the list result with pagination.
type ListFilesResult struct {
	Items      []FileDTO
	NextCursor *string
	HasMore    bool
}

// ListFilesHandler handles list files query.
type ListFilesHandler func(ctx context.Context, query ListFilesQuery) (*ListFilesResult, error)

// GetDownloadURLQuery generates a presigned download URL.
type GetDownloadURLQuery struct {
	FileID    string
	ExpiresIn int // seconds
}

// GetDownloadURLResult contains the presigned download URL.
type GetDownloadURLResult struct {
	FileID      string
	OwnerID     *string
	Filename    string
	Size        int64
	DownloadURL string
	ExpiresAt   string
}

// GetDownloadURLHandler handles get download URL query.
type GetDownloadURLHandler func(ctx context.Context, query GetDownloadURLQuery) (*GetDownloadURLResult, error)

// GetProcessingStatusQuery retrieves processing status by ID.
type GetProcessingStatusQuery struct {
	ProcessingID string
}

// GetProcessingStatusHandler handles get processing status query.
type GetProcessingStatusHandler func(ctx context.Context, query GetProcessingStatusQuery) (*ProcessingDTO, error)

// ListProcessingsQuery lists processings for a file.
type ListProcessingsQuery struct {
	FileID string
	Status *domain.ProcessingStatus
	Limit  int
	Cursor *string
}

// ListProcessingsResult contains the list result with pagination.
type ListProcessingsResult struct {
	Items      []ProcessingDTO
	NextCursor *string
	HasMore    bool
}

// ListProcessingsHandler handles list processings query.
type ListProcessingsHandler func(ctx context.Context, query ListProcessingsQuery) (*ListProcessingsResult, error)

// FileDTO is a data transfer object for file information.
type FileDTO struct {
	ID              string           `json:"id"`
	OwnerID         *string          `json:"owner_id,omitempty"`
	Filename        string           `json:"filename"`
	StoredName      string           `json:"stored_name"`
	MimeType        string           `json:"mime_type"`
	Size            int64            `json:"size"`
	Path            string           `json:"path"`
	StorageProvider string           `json:"storage_provider"`
	Checksum        string           `json:"checksum"`
	Metadata        *FileMetadataDTO `json:"metadata,omitempty"`
	AccessLevel     string           `json:"access_level"`
	UploadedAt      string           `json:"uploaded_at"`
	ExpiresAt       *string          `json:"expires_at,omitempty"`
	ProcessedAt     *string          `json:"processed_at,omitempty"`
	CreatedAt       string           `json:"created_at"`
	UpdatedAt       string           `json:"updated_at"`
}

// FileMetadataDTO is a DTO for file metadata.
type FileMetadataDTO struct {
	Width      *int              `json:"width,omitempty"`
	Height     *int              `json:"height,omitempty"`
	Duration   *int              `json:"duration,omitempty"`
	Pages      *int              `json:"pages,omitempty"`
	Thumbnail  *string           `json:"thumbnail,omitempty"`
	OriginalID *string           `json:"original_id,omitempty"`
	Custom     map[string]string `json:"custom,omitempty"`
}

// ProcessingDTO is a DTO for processing information.
type ProcessingDTO struct {
	ID           string  `json:"id"`
	FileID       string  `json:"file_id"`
	Operation    string  `json:"operation"`
	Status       string  `json:"status"`
	ResultFileID *string `json:"result_file_id,omitempty"`
	Error        *string `json:"error,omitempty"`
	Duration     *string `json:"duration,omitempty"`
	StartedAt    *string `json:"started_at,omitempty"`
	CompletedAt  *string `json:"completed_at,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

// FileStatsDTO contains file statistics.
type FileStatsDTO struct {
	TotalFiles    int64            `json:"total_files"`
	TotalSize     int64            `json:"total_size"`
	ByMimeType    map[string]int64 `json:"by_mime_type"`
	ByAccessLevel map[string]int64 `json:"by_access_level"`
	ByStorageType map[string]int64 `json:"by_storage_type"`
}

// GetFileStatsQuery retrieves file statistics.
type GetFileStatsQuery struct {
	OwnerID *string
}

// GetFileStatsHandler handles get file stats query.
type GetFileStatsHandler func(ctx context.Context, query GetFileStatsQuery) (*FileStatsDTO, error)

// ToFileDTO converts a File domain entity to DTO.
func ToFileDTO(file *domain.File) *FileDTO {
	dto := &FileDTO{
		ID:              file.ID().String(),
		Filename:        file.Filename(),
		StoredName:      file.StoredName(),
		MimeType:        file.MimeType(),
		Size:            file.Size(),
		Path:            file.Path(),
		StorageProvider: string(file.StorageProvider()),
		Checksum:        file.Checksum(),
		AccessLevel:     string(file.AccessLevel()),
		UploadedAt:      file.UploadedAt().Format(time.RFC3339),
		CreatedAt:       file.CreatedAt().Format(time.RFC3339),
		UpdatedAt:       file.UpdatedAt().Format(time.RFC3339),
	}

	if file.OwnerID() != nil {
		ownerID := file.OwnerID().String()
		dto.OwnerID = &ownerID
	}

	if file.ExpiresAt() != nil {
		expiresAt := file.ExpiresAt().Format(time.RFC3339)
		dto.ExpiresAt = &expiresAt
	}

	if file.ProcessedAt() != nil {
		processedAt := file.ProcessedAt().Format(time.RFC3339)
		dto.ProcessedAt = &processedAt
	}

	metadata := file.Metadata()
	if metadata.Width != nil || metadata.Height != nil || metadata.Duration != nil ||
		metadata.Pages != nil || metadata.Thumbnail != nil || metadata.OriginalID != nil ||
		len(metadata.Custom) > 0 {

		dto.Metadata = &FileMetadataDTO{
			Width:    metadata.Width,
			Height:   metadata.Height,
			Duration: metadata.Duration,
			Pages:    metadata.Pages,
			Custom:   metadata.Custom,
		}

		if metadata.Thumbnail != nil {
			thumbnail := metadata.Thumbnail.String()
			dto.Metadata.Thumbnail = &thumbnail
		}

		if metadata.OriginalID != nil {
			originalID := metadata.OriginalID.String()
			dto.Metadata.OriginalID = &originalID
		}
	}

	return dto
}

// ToProcessingDTO converts a FileProcessing domain entity to DTO.
func ToProcessingDTO(processing *domain.FileProcessing) *ProcessingDTO {
	dto := &ProcessingDTO{
		ID:        processing.ID().String(),
		FileID:    processing.FileID().String(),
		Operation: string(processing.Operation()),
		Status:    string(processing.Status()),
		CreatedAt: processing.CreatedAt().Format(time.RFC3339),
	}

	if processing.ResultFileID() != nil {
		resultID := processing.ResultFileID().String()
		dto.ResultFileID = &resultID
	}

	if processing.Error() != nil {
		dto.Error = processing.Error()
	}

	if processing.Duration() != nil {
		duration := processing.Duration().String()
		dto.Duration = &duration
	}

	if processing.StartedAt() != nil {
		startedAt := processing.StartedAt().Format(time.RFC3339)
		dto.StartedAt = &startedAt
	}

	if processing.CompletedAt() != nil {
		completedAt := processing.CompletedAt().Format(time.RFC3339)
		dto.CompletedAt = &completedAt
	}

	return dto
}

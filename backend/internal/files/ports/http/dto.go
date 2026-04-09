package http

import (
	"time"

	"github.com/basilex/skeleton/internal/files/domain"
)

// UploadFileRequest represents a file upload request.
type UploadFileRequest struct {
	OwnerID         *string `json:"owner_id,omitempty"`
	Filename        string  `json:"filename" binding:"required"`
	Content         []byte  `json:"content" binding:"required"`
	MimeType        string  `json:"mime_type" binding:"required"`
	Size            int64   `json:"size" binding:"required"`
	StorageProvider string  `json:"storage_provider" binding:"required"`
	AccessLevel     string  `json:"access_level" binding:"required"`
}

// UploadFileResponse represents a file upload response.
type UploadFileResponse struct {
	FileID      string `json:"id"`
	StoredName  string `json:"stored_name"`
	StoragePath string `json:"storage_path"`
	Checksum    string `json:"checksum"`
	UploadedAt  string `json:"uploaded_at"`
}

// UploadURLRequest represents a request for presigned upload URL.
type UploadURLRequest struct {
	OwnerID         *string `json:"owner_id,omitempty"`
	Filename        string  `json:"filename" binding:"required"`
	MimeType        string  `json:"mime_type" binding:"required"`
	Size            int64   `json:"size" binding:"required"`
	StorageProvider string  `json:"storage_provider" binding:"required"`
	AccessLevel     string  `json:"access_level" binding:"required"`
	TTL             int     `json:"ttl"` // Time-to-live in seconds
}

// UploadURLResponse represents a presigned upload URL response.
type UploadURLResponse struct {
	UploadID  string            `json:"upload_id"`
	UploadURL string            `json:"upload_url"`
	Fields    map[string]string `json:"fields,omitempty"`
	ExpiresAt string            `json:"expires_at"`
}

// ConfirmUploadRequest represents an upload confirmation request.
type ConfirmUploadRequest struct {
	UploadID string `json:"upload_id" binding:"required"`
	Checksum string `json:"checksum"`
}

// ConfirmUploadResponse represents an upload confirmation response.
type ConfirmUploadResponse struct {
	FileID      string `json:"id"`
	Filename    string `json:"filename"`
	StoredName  string `json:"stored_name"`
	StoragePath string `json:"storage_path"`
	UploadedAt  string `json:"uploaded_at"`
}

// ProcessFileRequest represents a file processing request.
type ProcessFileRequest struct {
	FileID    string                 `json:"file_id" binding:"required"`
	Operation string                 `json:"operation" binding:"required"`
	Options   map[string]interface{} `json:"options"`
}

// ProcessFileResponse represents a processing request response.
type ProcessFileResponse struct {
	ProcessingID string `json:"processing_id"`
	Status       string `json:"status"`
}

// FileResponse represents a file response.
type FileResponse struct {
	ID              string                 `json:"id"`
	OwnerID         *string                `json:"owner_id,omitempty"`
	Filename        string                 `json:"filename"`
	StoredName      string                 `json:"stored_name"`
	MimeType        string                 `json:"mime_type"`
	Size            int64                  `json:"size"`
	Path            string                 `json:"path"`
	StorageProvider string                 `json:"storage_provider"`
	Checksum        string                 `json:"checksum"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	AccessLevel     string                 `json:"access_level"`
	UploadedAt      string                 `json:"uploaded_at"`
	ExpiresAt       *string                `json:"expires_at,omitempty"`
	ProcessedAt     *string                `json:"processed_at,omitempty"`
	CreatedAt       string                 `json:"created_at"`
	UpdatedAt       string                 `json:"updated_at"`
}

// ProcessingResponse represents a processing response.
type ProcessingResponse struct {
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

// ListFilesRequest represents a list files request.
type ListFilesRequest struct {
	OwnerID      *string `form:"owner_id"`
	MimeType     *string `form:"mime_type"`
	AccessLevel  *string `form:"access_level"`
	UploadedFrom *string `form:"uploaded_from"`
	UploadedTo   *string `form:"uploaded_to"`
	Limit        int     `form:"limit"`
	Cursor       *string `form:"cursor"`
}

// ListFilesResponse represents a list files response.
type ListFilesResponse struct {
	Items      []FileResponse `json:"items"`
	NextCursor *string        `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

// ListProcessingsRequest represents a list processings request.
type ListProcessingsRequest struct {
	FileID string  `form:"file_id" binding:"required"`
	Status *string `form:"status"`
	Limit  int     `form:"limit"`
	Cursor *string `form:"cursor"`
}

// ListProcessingsResponse represents a list processings response.
type ListProcessingsResponse struct {
	Items      []ProcessingResponse `json:"items"`
	NextCursor *string              `json:"next_cursor,omitempty"`
	HasMore    bool                 `json:"has_more"`
}

// DownloadURLResponse represents a download URL response.
type DownloadURLResponse struct {
	FileID      string  `json:"file_id"`
	OwnerID     *string `json:"owner_id,omitempty"`
	Filename    string  `json:"filename"`
	Size        int64   `json:"size"`
	DownloadURL string  `json:"download_url"`
	ExpiresAt   string  `json:"expires_at"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// ToFileResponse converts a File domain entity to FileResponse.
func ToFileResponse(file *domain.File) *FileResponse {
	resp := &FileResponse{
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
		resp.OwnerID = &ownerID
	}

	if file.ExpiresAt() != nil {
		expiresAt := file.ExpiresAt().Format(time.RFC3339)
		resp.ExpiresAt = &expiresAt
	}

	if file.ProcessedAt() != nil {
		processedAt := file.ProcessedAt().Format(time.RFC3339)
		resp.ProcessedAt = &processedAt
	}

	metadata := file.Metadata()
	if len(metadata.Custom) > 0 || metadata.Width != nil || metadata.Height != nil {
		resp.Metadata = make(map[string]interface{})
		if metadata.Width != nil {
			resp.Metadata["width"] = *metadata.Width
		}
		if metadata.Height != nil {
			resp.Metadata["height"] = *metadata.Height
		}
		for k, v := range metadata.Custom {
			resp.Metadata[k] = v
		}
	}

	return resp
}

// ToProcessingResponse converts a FileProcessing domain entity to ProcessingResponse.
func ToProcessingResponse(processing *domain.FileProcessing) *ProcessingResponse {
	resp := &ProcessingResponse{
		ID:        processing.ID().String(),
		FileID:    processing.FileID().String(),
		Operation: string(processing.Operation()),
		Status:    string(processing.Status()),
		CreatedAt: processing.CreatedAt().Format(time.RFC3339),
	}

	if processing.ResultFileID() != nil {
		resultID := processing.ResultFileID().String()
		resp.ResultFileID = &resultID
	}

	if processing.Error() != nil {
		errMsg := *processing.Error()
		resp.Error = &errMsg
	}

	if processing.Duration() != nil {
		duration := processing.Duration().String()
		resp.Duration = &duration
	}

	if processing.StartedAt() != nil {
		startedAt := processing.StartedAt().Format(time.RFC3339)
		resp.StartedAt = &startedAt
	}

	if processing.CompletedAt() != nil {
		completedAt := processing.CompletedAt().Format(time.RFC3339)
		resp.CompletedAt = &completedAt
	}

	return resp
}

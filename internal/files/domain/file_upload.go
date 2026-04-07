package domain

import (
	"time"
)

// UploadStatus represents the status of a file upload.
type UploadStatus string

const (
	UploadPending   UploadStatus = "pending"   // Upload URL generated, waiting for upload
	UploadCompleted UploadStatus = "completed" // Upload finished successfully
	UploadFailed    UploadStatus = "failed"    // Upload failed
	UploadExpired   UploadStatus = "expired"   // Upload URL expired
)

// FileUpload represents a presigned upload URL for large file uploads.
// This aggregate is used for uploads bigger than 5MB where direct upload is not suitable.
// Workflow:
//  1. User requests upload URL → FileUpload created with presigned URL
//  2. User uploads directly to storage using the URL
//  3. User confirms upload → FileUpload marked as completed, File created
type FileUpload struct {
	id        UploadID
	file      *File             // File metadata (created after confirmation)
	uploadURL string            // Presigned URL for upload
	fields    map[string]string // Additional form fields for upload
	status    UploadStatus
	expiresAt time.Time // URL expiration time
	createdAt time.Time
}

// NewFileUpload creates a new FileUpload for presigned URL upload.
// The upload URL will expire after the specified TTL.
func NewFileUpload(file *File, ttl time.Duration) (*FileUpload, error) {
	if file == nil {
		return nil, NewValidationError("file", "file cannot be nil")
	}
	if ttl <= 0 {
		return nil, NewValidationError("ttl", "TTL must be positive")
	}

	now := time.Now()
	return &FileUpload{
		id:        NewUploadID(),
		file:      file,
		status:    UploadPending,
		expiresAt: now.Add(ttl),
		createdAt: now,
		fields:    make(map[string]string),
	}, nil
}

// ReconstituteFileUpload reconstructs a FileUpload from persistence.
func ReconstituteFileUpload(
	id UploadID,
	file *File,
	uploadURL string,
	fields map[string]string,
	status UploadStatus,
	expiresAt time.Time,
	createdAt time.Time,
) *FileUpload {
	if fields == nil {
		fields = make(map[string]string)
	}
	return &FileUpload{
		id:        id,
		file:      file,
		uploadURL: uploadURL,
		fields:    fields,
		status:    status,
		expiresAt: expiresAt,
		createdAt: createdAt,
	}
}

// ID returns the upload ID.
func (u *FileUpload) ID() UploadID { return u.id }

// File returns the file metadata.
func (u *FileUpload) File() *File { return u.file }

// UploadURL returns the presigned upload URL.
func (u *FileUpload) UploadURL() string { return u.uploadURL }

// Fields returns additional form fields for upload.
func (u *FileUpload) Fields() map[string]string {
	if u.fields == nil {
		return make(map[string]string)
	}
	// Return a copy to prevent mutation
	fields := make(map[string]string, len(u.fields))
	for k, v := range u.fields {
		fields[k] = v
	}
	return fields
}

// Status returns the upload status.
func (u *FileUpload) Status() UploadStatus { return u.status }

// ExpiresAt returns the expiration time.
func (u *FileUpload) ExpiresAt() time.Time { return u.expiresAt }

// CreatedAt returns the creation time.
func (u *FileUpload) CreatedAt() time.Time { return u.createdAt }

// SetUploadURL sets the presigned upload URL and additional fields.
// Called by storage provider after generating the presigned URL.
func (u *FileUpload) SetUploadURL(url string, fields map[string]string) {
	u.uploadURL = url
	if fields != nil {
		u.fields = fields
	}
}

// MarkCompleted marks the upload as completed.
// Called after user confirms successful upload.
func (u *FileUpload) MarkCompleted() error {
	if u.status == UploadCompleted {
		return ErrUploadAlreadyCompleted
	}
	if u.status == UploadExpired {
		return ErrUploadExpired
	}
	if u.status == UploadFailed {
		return ErrUploadNotCompleted
	}

	u.status = UploadCompleted
	return nil
}

// MarkFailed marks the upload as failed.
func (u *FileUpload) MarkFailed() {
	u.status = UploadFailed
}

// IsExpired returns true if the upload URL has expired.
func (u *FileUpload) IsExpired() bool {
	return time.Now().After(u.expiresAt)
}

// CanUpload returns true if the upload can be performed.
func (u *FileUpload) CanUpload() bool {
	return u.status == UploadPending && !u.IsExpired()
}

// IsValidStatus validates the upload status.
func IsValidStatus(status UploadStatus) bool {
	switch status {
	case UploadPending, UploadCompleted, UploadFailed, UploadExpired:
		return true
	default:
		return false
	}
}

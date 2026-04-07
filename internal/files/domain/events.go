package domain

import (
	"time"
)

// FileUploaded event is published when a file is successfully uploaded.
type FileUploaded struct {
	FileID      FileID
	OwnerID     *string
	Filename    string
	MimeType    string
	Size        int64
	StoragePath string
	AccessLevel AccessLevel
	UploadedAt  time.Time
}

// FileDownloaded event is published when a file is downloaded.
type FileDownloaded struct {
	FileID       FileID
	UserID       *string
	AccessType   string // "view" or "download"
	DownloadedAt time.Time
}

// FileDeleted event is published when a file is deleted.
type FileDeleted struct {
	FileID    FileID
	OwnerID   *string
	DeletedAt time.Time
}

// FileProcessed event is published when file processing completes.
type FileProcessed struct {
	ProcessingID ProcessingID
	FileID       FileID
	ResultFileID FileID
	Operation    ProcessingOperation
	Status       ProcessingStatus
	CompletedAt  time.Time
}

// FileUploaded event name.
func (e FileUploaded) EventName() string {
	return "files.file_uploaded"
}

// FileDownloaded event name.
func (e FileDownloaded) EventName() string {
	return "files.file_downloaded"
}

// FileDeleted event name.
func (e FileDeleted) EventName() string {
	return "files.file_deleted"
}

// FileProcessed event name.
func (e FileProcessed) EventName() string {
	return "files.file_processed"
}

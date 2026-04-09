package domain

import (
	"github.com/basilex/skeleton/pkg/uuid"
)

// FileID represents a unique identifier for a File aggregate.
type FileID uuid.UUID

// NewFileID generates a new FileID using UUID v7.
func NewFileID() FileID {
	return FileID(uuid.NewV7())
}

// ParseFileID parses a string into a FileID.
// Returns an error if the string is empty or invalid.
func ParseFileID(s string) (FileID, error) {
	if s == "" {
		return FileID{}, ErrInvalidFileID
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return FileID{}, ErrInvalidFileID
	}
	return FileID(u), nil
}

// String returns the string representation of FileID.
func (id FileID) String() string {
	return uuid.UUID(id).String()
}

// UploadID represents a unique identifier for a FileUpload aggregate.
type UploadID uuid.UUID

// NewUploadID generates a new UploadID using UUID v7.
func NewUploadID() UploadID {
	return UploadID(uuid.NewV7())
}

// ParseUploadID parses a string into an UploadID.
// Returns an error if the string is empty or invalid.
func ParseUploadID(s string) (UploadID, error) {
	if s == "" {
		return UploadID{}, ErrInvalidUploadID
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return UploadID{}, ErrInvalidUploadID
	}
	return UploadID(u), nil
}

// String returns the string representation of UploadID.
func (id UploadID) String() string {
	return uuid.UUID(id).String()
}

// ProcessingID represents a unique identifier for a FileProcessing aggregate.
type ProcessingID uuid.UUID

// NewProcessingID generates a new ProcessingID using UUID v7.
func NewProcessingID() ProcessingID {
	return ProcessingID(uuid.NewV7())
}

// ParseProcessingID parses a string into a ProcessingID.
// Returns an error if the string is empty or invalid.
func ParseProcessingID(s string) (ProcessingID, error) {
	if s == "" {
		return ProcessingID{}, ErrInvalidProcessingID
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return ProcessingID{}, ErrInvalidProcessingID
	}
	return ProcessingID(u), nil
}

// String returns the string representation of ProcessingID.
func (id ProcessingID) String() string {
	return uuid.UUID(id).String()
}

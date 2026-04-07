package domain

import (
	"github.com/basilex/skeleton/pkg/uuid"
)

// FileID represents a unique identifier for a File aggregate.
type FileID string

// NewFileID generates a new FileID using UUID v7.
func NewFileID() FileID {
	return FileID(uuid.NewV7().String())
}

// ParseFileID parses a string into a FileID.
// Returns an error if the string is empty or invalid.
func ParseFileID(s string) (FileID, error) {
	if s == "" {
		return "", ErrInvalidFileID
	}
	return FileID(s), nil
}

// String returns the string representation of FileID.
func (id FileID) String() string {
	return string(id)
}

// UploadID represents a unique identifier for a FileUpload aggregate.
type UploadID string

// NewUploadID generates a new UploadID using UUID v7.
func NewUploadID() UploadID {
	return UploadID(uuid.NewV7().String())
}

// ParseUploadID parses a string into an UploadID.
// Returns an error if the string is empty or invalid.
func ParseUploadID(s string) (UploadID, error) {
	if s == "" {
		return "", ErrInvalidUploadID
	}
	return UploadID(s), nil
}

// String returns the string representation of UploadID.
func (id UploadID) String() string {
	return string(id)
}

// ProcessingID represents a unique identifier for a FileProcessing aggregate.
type ProcessingID string

// NewProcessingID generates a new ProcessingID using UUID v7.
func NewProcessingID() ProcessingID {
	return ProcessingID(uuid.NewV7().String())
}

// ParseProcessingID parses a string into a ProcessingID.
// Returns an error if the string is empty or invalid.
func ParseProcessingID(s string) (ProcessingID, error) {
	if s == "" {
		return "", ErrInvalidProcessingID
	}
	return ProcessingID(s), nil
}

// String returns the string representation of ProcessingID.
func (id ProcessingID) String() string {
	return string(id)
}

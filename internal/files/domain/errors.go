package domain

import (
	"errors"
	"fmt"
)

// File errors.
var (
	// ErrInvalidFileID indicates that the file ID is invalid or malformed.
	ErrInvalidFileID = errors.New("invalid file ID")

	// ErrFileNotFound indicates that the file was not found in the repository.
	ErrFileNotFound = errors.New("file not found")

	// ErrFileAlreadyExists indicates that a file with the same ID already exists.
	ErrFileAlreadyExists = errors.New("file already exists")

	// ErrInvalidFileState indicates that the file is in an invalid state for the operation.
	ErrInvalidFileState = errors.New("invalid file state")

	// ErrInvalidFileType indicates that the file type/MIME type is not allowed.
	ErrInvalidFileType = errors.New("invalid file type")

	// ErrFileTooLarge indicates that the file exceeds the maximum allowed size.
	ErrFileTooLarge = errors.New("file too large")

	// ErrFileExpired indicates that the file has expired and is no longer accessible.
	ErrFileExpired = errors.New("file has expired")

	// ErrFileAccessDenied indicates that the user does not have permission to access the file.
	ErrFileAccessDenied = errors.New("file access denied")

	// ErrInvalidChecksum indicates that the file checksum does not match.
	ErrInvalidChecksum = errors.New("invalid file checksum")

	// ErrFileProcessing indicates that file processing failed.
	ErrFileProcessing = errors.New("file processing failed")
)

// Upload errors.
var (
	// ErrInvalidUploadID indicates that the upload ID is invalid or malformed.
	ErrInvalidUploadID = errors.New("invalid upload ID")

	// ErrUploadNotFound indicates that the upload was not found.
	ErrUploadNotFound = errors.New("upload not found")

	// ErrUploadExpired indicates that the upload URL has expired.
	ErrUploadExpired = errors.New("upload URL has expired")

	// ErrUploadAlreadyCompleted indicates that the upload has already been completed.
	ErrUploadAlreadyCompleted = errors.New("upload already completed")

	// ErrUploadNotCompleted indicates that the upload has not been completed yet.
	ErrUploadNotCompleted = errors.New("upload not completed")
)

// Processing errors.
var (
	// ErrInvalidProcessingID indicates that the processing ID is invalid.
	ErrInvalidProcessingID = errors.New("invalid processing ID")

	// ErrProcessingNotFound indicates that the processing record was not found.
	ErrProcessingNotFound = errors.New("processing not found")

	// ErrProcessingAlreadyRunning indicates that processing is already running for this file.
	ErrProcessingAlreadyRunning = errors.New("processing already running")

	// ErrProcessingFailed indicates that file processing failed.
	ErrProcessingFailed = errors.New("processing failed")

	// ErrInvalidProcessingOperation indicates that the processing operation is not supported.
	ErrInvalidProcessingOperation = errors.New("invalid processing operation")
)

// Storage errors.
var (
	// ErrStorageProvider indicates that the storage provider is not supported.
	ErrStorageProvider = errors.New("unsupported storage provider")

	// ErrStorageFailed indicates that a storage operation failed.
	ErrStorageFailed = errors.New("storage operation failed")

	// ErrFileNotFoundInStorage indicates that the file was not found in storage.
	ErrFileNotFoundInStorage = errors.New("file not found in storage")
)

// ValidationError represents a validation error with details.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

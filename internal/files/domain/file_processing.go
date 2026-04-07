package domain

import (
	"fmt"
	"time"
)

// ProcessingOperation represents a file processing operation type.
type ProcessingOperation string

const (
	OperationResize    ProcessingOperation = "resize"    // Resize to specific dimensions
	OperationCrop      ProcessingOperation = "crop"      // Crop to specific area
	OperationCompress  ProcessingOperation = "compress"  // Compress file size
	OperationConvert   ProcessingOperation = "convert"   // Convert to different format
	OperationThumbnail ProcessingOperation = "thumbnail" // Generate thumbnail
	OperationWatermark ProcessingOperation = "watermark" // Add watermark
)

// ProcessingStatus represents the status of file processing.
type ProcessingStatus string

const (
	ProcessingPending   ProcessingStatus = "pending"   // Waiting to be processed
	ProcessingRunning   ProcessingStatus = "running"   // Currently being processed
	ProcessingCompleted ProcessingStatus = "completed" // Processing finished successfully
	ProcessingFailed    ProcessingStatus = "failed"    // Processing failed
)

// ProcessingOptions contains options for file processing operations.
type ProcessingOptions struct {
	Width     *int              // Target width (for resize/crop)
	Height    *int              // Target height (for resize/crop)
	X         *int              // Crop offset X
	Y         *int              // Crop offset Y
	Quality   *int              // Quality (for compress, 1-100)
	Format    *string           // Target format (for convert)
	Watermark *string           // Watermark text or path
	Custom    map[string]string // Custom options
}

// FileProcessing represents a file processing task.
// This aggregate tracks the status of asynchronous file processing operations
// such as image resize, video transcoding, or document conversion.
// Processing is performed by the Tasks context.
type FileProcessing struct {
	id           ProcessingID
	fileID       FileID
	operation    ProcessingOperation
	options      ProcessingOptions
	status       ProcessingStatus
	resultFileID *FileID // Resulting processed file ID (if successful)
	error        *string // Error message (if failed)
	startedAt    *time.Time
	completedAt  *time.Time
	createdAt    time.Time
}

// NewFileProcessing creates a new file processing task.
func NewFileProcessing(
	fileID FileID,
	operation ProcessingOperation,
	options ProcessingOptions,
) (*FileProcessing, error) {
	if !isValidOperation(operation) {
		return nil, NewValidationError("operation", "invalid processing operation")
	}

	return &FileProcessing{
		id:        NewProcessingID(),
		fileID:    fileID,
		operation: operation,
		options:   options,
		status:    ProcessingPending,
		createdAt: time.Now(),
	}, nil
}

// ReconstituteFileProcessing reconstructs a FileProcessing from persistence.
func ReconstituteFileProcessing(
	id ProcessingID,
	fileID FileID,
	operation ProcessingOperation,
	options ProcessingOptions,
	status ProcessingStatus,
	resultFileID *FileID,
	error *string,
	startedAt *time.Time,
	completedAt *time.Time,
	createdAt time.Time,
) *FileProcessing {
	return &FileProcessing{
		id:           id,
		fileID:       fileID,
		operation:    operation,
		options:      options,
		status:       status,
		resultFileID: resultFileID,
		error:        error,
		startedAt:    startedAt,
		completedAt:  completedAt,
		createdAt:    createdAt,
	}
}

// ID returns the processing ID.
func (p *FileProcessing) ID() ProcessingID { return p.id }

// FileID returns the source file ID.
func (p *FileProcessing) FileID() FileID { return p.fileID }

// Operation returns the processing operation type.
func (p *FileProcessing) Operation() ProcessingOperation { return p.operation }

// Options returns the processing options.
func (p *FileProcessing) Options() ProcessingOptions { return p.options }

// Status returns the processing status.
func (p *FileProcessing) Status() ProcessingStatus { return p.status }

// ResultFileID returns the result file ID (if successful).
func (p *FileProcessing) ResultFileID() *FileID { return p.resultFileID }

// Error returns the error message (if failed).
func (p *FileProcessing) Error() *string { return p.error }

// StartedAt returns the processing start time.
func (p *FileProcessing) StartedAt() *time.Time { return p.startedAt }

// CompletedAt returns the processing completion time.
func (p *FileProcessing) CompletedAt() *time.Time { return p.completedAt }

// CreatedAt returns the creation time.
func (p *FileProcessing) CreatedAt() time.Time { return p.createdAt }

// Start marks the processing as started.
// Called by the task worker when processing begins.
func (p *FileProcessing) Start() error {
	if p.status != ProcessingPending {
		return fmt.Errorf("%w: cannot start processing in %s status",
			ErrInvalidFileState, p.status)
	}

	now := time.Now()
	p.status = ProcessingRunning
	p.startedAt = &now
	return nil
}

// Complete marks the processing as completed successfully.
// resultFileID is the ID of the processed file.
func (p *FileProcessing) Complete(resultFileID FileID) error {
	if p.status != ProcessingRunning {
		return fmt.Errorf("%w: cannot complete processing in %s status",
			ErrInvalidFileState, p.status)
	}

	now := time.Now()
	p.status = ProcessingCompleted
	p.resultFileID = &resultFileID
	p.completedAt = &now
	return nil
}

// Fail marks the processing as failed with an error message.
func (p *FileProcessing) Fail(errorMessage string) error {
	if p.status != ProcessingRunning {
		return fmt.Errorf("%w: cannot fail processing in %s status",
			ErrInvalidFileState, p.status)
	}

	now := time.Now()
	p.status = ProcessingFailed
	p.error = &errorMessage
	p.completedAt = &now
	return nil
}

// IsCompleted returns true if processing is completed successfully.
func (p *FileProcessing) IsCompleted() bool {
	return p.status == ProcessingCompleted
}

// IsFailed returns true if processing failed.
func (p *FileProcessing) IsFailed() bool {
	return p.status == ProcessingFailed
}

// IsRunning returns true if processing is currently running.
func (p *FileProcessing) IsRunning() bool {
	return p.status == ProcessingRunning
}

// IsPending returns true if processing is pending.
func (p *FileProcessing) IsPending() bool {
	return p.status == ProcessingPending
}

// Duration returns the processing duration (if completed).
func (p *FileProcessing) Duration() *time.Duration {
	if p.startedAt == nil || p.completedAt == nil {
		return nil
	}

	duration := p.completedAt.Sub(*p.startedAt)
	return &duration
}

// isValidOperation validates the processing operation.
func isValidOperation(op ProcessingOperation) bool {
	switch op {
	case OperationResize, OperationCrop, OperationCompress,
		OperationConvert, OperationThumbnail, OperationWatermark:
		return true
	default:
		return false
	}
}

// String returns the string representation of ProcessingOperation.
func (op ProcessingOperation) String() string {
	return string(op)
}

// String returns the string representation of ProcessingStatus.
func (s ProcessingStatus) String() string {
	return string(s)
}

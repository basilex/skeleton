package command

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/basilex/skeleton/internal/files/domain"
	identitydomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

// UploadFileCommand uploads a file directly (for small files <5MB).
type UploadFileCommand struct {
	OwnerID         *string
	Filename        string
	Content         io.Reader
	MimeType        string
	Size            int64
	StorageProvider domain.StorageProvider
	AccessLevel     domain.AccessLevel
}

// UploadFileResult contains the result of file upload.
type UploadFileResult struct {
	FileID      string
	StoredName  string
	StoragePath string
	Checksum    string
	UploadedAt  string
}

// UploadFileHandler handles file upload command.
type UploadFileHandler func(ctx context.Context, cmd UploadFileCommand) (*UploadFileResult, error)

// DeleteFileCommand deletes a file from storage and repository.
type DeleteFileCommand struct {
	FileID string
}

// DeleteFileHandler handles file deletion command.
type DeleteFileHandler func(ctx context.Context, cmd DeleteFileCommand) error

// RequestUploadURLCommand requests a presigned upload URL for large files.
type RequestUploadURLCommand struct {
	OwnerID         *string
	Filename        string
	MimeType        string
	Size            int64
	StorageProvider domain.StorageProvider
	AccessLevel     domain.AccessLevel
	TTL             int // URL time-to-live in seconds
}

// RequestUploadURLResult contains the presigned upload URL.
type RequestUploadURLResult struct {
	UploadID  string
	UploadURL string
	Fields    map[string]string
	ExpiresAt string
}

// RequestUploadURLHandler handles upload URL request command.
type RequestUploadURLHandler func(ctx context.Context, cmd RequestUploadURLCommand) (*RequestUploadURLResult, error)

// ConfirmUploadCommand confirms that a file has been uploaded to storage.
type ConfirmUploadCommand struct {
	UploadID string
	Checksum string
}

// ConfirmUploadResult contains the confirmed file information.
type ConfirmUploadResult struct {
	FileID      string
	Filename    string
	StoredName  string
	StoragePath string
	UploadedAt  string
}

// ConfirmUploadHandler handles upload confirmation command.
type ConfirmUploadHandler func(ctx context.Context, cmd ConfirmUploadCommand) (*ConfirmUploadResult, error)

// RequestProcessingCommand requests file processing.
type RequestProcessingCommand struct {
	FileID    string
	Operation domain.ProcessingOperation
	Options   domain.ProcessingOptions
}

// RequestProcessingResult contains the processing task information.
type RequestProcessingResult struct {
	ProcessingID string
	Status       string
}

// RequestProcessingHandler handles processing request command.
type RequestProcessingHandler func(ctx context.Context, cmd RequestProcessingCommand) (*RequestProcessingResult, error)

// CleanupExpiredFilesCommand removes expired files.
type CleanupExpiredFilesCommand struct {
	BatchSize int
}

// CleanupExpiredFilesHandler handles file cleanup command.
type CleanupExpiredFilesHandler func(ctx context.Context, cmd CleanupExpiredFilesCommand) (int64, error)

// StorageProvider interface for file storage operations.
type StorageProvider interface {
	Upload(ctx context.Context, path string, reader io.Reader, contentType string) error
	Delete(ctx context.Context, path string) error
}

// validate validates the upload file command.
func (cmd UploadFileCommand) validate() error {
	if cmd.Filename == "" {
		return domain.NewValidationError("filename", "filename cannot be empty")
	}
	if cmd.Content == nil {
		return domain.NewValidationError("content", "content cannot be nil")
	}
	if cmd.MimeType == "" {
		return domain.NewValidationError("mimeType", "MIME type cannot be empty")
	}
	if cmd.Size < 0 {
		return domain.NewValidationError("size", "file size cannot be negative")
	}
	if !isValidStorageProvider(cmd.StorageProvider) {
		return domain.NewValidationError("storageProvider", "invalid storage provider")
	}
	if !isValidAccessLevel(cmd.AccessLevel) {
		return domain.NewValidationError("accessLevel", "invalid access level")
	}
	return nil
}

// validate validates the request upload URL command.
func (cmd RequestUploadURLCommand) validate() error {
	if cmd.Filename == "" {
		return domain.NewValidationError("filename", "filename cannot be empty")
	}
	if cmd.MimeType == "" {
		return domain.NewValidationError("mimeType", "MIME type cannot be empty")
	}
	if cmd.Size < 0 {
		return domain.NewValidationError("size", "file size cannot be negative")
	}
	if cmd.TTL <= 0 {
		return domain.NewValidationError("ttl", "TTL must be positive")
	}
	if !isValidStorageProvider(cmd.StorageProvider) {
		return domain.NewValidationError("storageProvider", "invalid storage provider")
	}
	if !isValidAccessLevel(cmd.AccessLevel) {
		return domain.NewValidationError("accessLevel", "invalid access level")
	}
	return nil
}

// validate validates the processing request command.
func (cmd RequestProcessingCommand) validate() error {
	if cmd.FileID == "" {
		return domain.NewValidationError("fileID", "file ID cannot be empty")
	}
	if !isValidProcessingOperation(cmd.Operation) {
		return domain.NewValidationError("operation", "invalid processing operation")
	}
	return nil
}

func isValidStorageProvider(p domain.StorageProvider) bool {
	switch p {
	case domain.StorageS3, domain.StorageGCS, domain.StorageLocal:
		return true
	default:
		return false
	}
}

func isValidAccessLevel(a domain.AccessLevel) bool {
	switch a {
	case domain.AccessPublic, domain.AccessPrivate, domain.AccessRestricted:
		return true
	default:
		return false
	}
}

func isValidProcessingOperation(o domain.ProcessingOperation) bool {
	switch o {
	case domain.OperationResize, domain.OperationCrop, domain.OperationCompress,
		domain.OperationConvert, domain.OperationThumbnail, domain.OperationWatermark:
		return true
	default:
		return false
	}
}

// Helper functions
func NewFileID() string {
	return domain.NewFileID().String()
}

func ParseFileID(s string) (domain.FileID, error) {
	return domain.ParseFileID(s)
}

// UploadFileHandler implementation
type uploadFileHandler struct {
	fileRepo domain.FileRepository
	storage  StorageProvider
	eventBus eventbus.Bus
}

func NewUploadFileHandler(
	fileRepo domain.FileRepository,
	storage StorageProvider,
	eventBus eventbus.Bus,
) UploadFileHandler {
	h := &uploadFileHandler{
		fileRepo: fileRepo,
		storage:  storage,
		eventBus: eventBus,
	}
	return h.Upload
}

func (h *uploadFileHandler) Upload(ctx context.Context, cmd UploadFileCommand) (*UploadFileResult, error) {
	if err := cmd.validate(); err != nil {
		return nil, err
	}

	var ownerID *identitydomain.UserID
	if cmd.OwnerID != nil {
		uid, err := identitydomain.ParseUserID(*cmd.OwnerID)
		if err != nil {
			return nil, fmt.Errorf("invalid owner ID: %w", err)
		}
		ownerID = &uid
	}

	file, err := domain.NewFile(
		ownerID,
		cmd.Filename,
		cmd.MimeType,
		cmd.Size,
		cmd.StorageProvider,
		cmd.AccessLevel,
	)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}

	storagePath := file.GeneratePath()
	if err := file.SetPath(storagePath); err != nil {
		return nil, fmt.Errorf("set path: %w", err)
	}

	if err := h.storage.Upload(ctx, storagePath, cmd.Content, cmd.MimeType); err != nil {
		return nil, fmt.Errorf("upload to storage: %w", err)
	}

	if err := h.fileRepo.Create(ctx, file); err != nil {
		_ = h.storage.Delete(ctx, storagePath)
		return nil, fmt.Errorf("save file: %w", err)
	}

	var ownerIDStr *string
	if file.OwnerID() != nil {
		s := file.OwnerID().String()
		ownerIDStr = &s
	}

	event := domain.FileUploaded{
		FileID:      file.ID(),
		OwnerID:     ownerIDStr,
		Filename:    file.Filename(),
		MimeType:    file.MimeType(),
		Size:        file.Size(),
		StoragePath: file.Path(),
		AccessLevel: file.AccessLevel(),
		UploadedAt:  file.UploadedAt(),
	}
	_ = h.eventBus.Publish(ctx, event)

	return &UploadFileResult{
		FileID:      file.ID().String(),
		StoredName:  file.StoredName(),
		StoragePath: file.Path(),
		Checksum:    file.Checksum(),
		UploadedAt:  file.UploadedAt().Format(time.RFC3339),
	}, nil
}

// DeleteFileHandler implementation
type deleteFileHandler struct {
	fileRepo domain.FileRepository
	storage  StorageProvider
	eventBus eventbus.Bus
}

func NewDeleteFileHandler(
	fileRepo domain.FileRepository,
	storage StorageProvider,
	eventBus eventbus.Bus,
) DeleteFileHandler {
	h := &deleteFileHandler{
		fileRepo: fileRepo,
		storage:  storage,
		eventBus: eventBus,
	}
	return h.Delete
}

func (h *deleteFileHandler) Delete(ctx context.Context, cmd DeleteFileCommand) error {
	fileID, err := ParseFileID(cmd.FileID)
	if err != nil {
		return err
	}

	file, err := h.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("get file: %w", err)
	}

	_ = h.storage.Delete(ctx, file.Path())

	if err := h.fileRepo.Delete(ctx, fileID); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	var ownerID *string
	if file.OwnerID() != nil {
		s := file.OwnerID().String()
		ownerID = &s
	}

	event := domain.FileDeleted{
		FileID:    fileID,
		OwnerID:   ownerID,
		DeletedAt: time.Now(),
	}
	_ = h.eventBus.Publish(ctx, event)

	return nil
}

// RequestUploadURLHandler implementation
type requestUploadURLHandler struct {
	uploadRepo domain.UploadRepository
	fileRepo   domain.FileRepository
	storage    StorageProvider
}

func NewRequestUploadURLHandler(
	uploadRepo domain.UploadRepository,
	fileRepo domain.FileRepository,
	storage StorageProvider,
) RequestUploadURLHandler {
	h := &requestUploadURLHandler{
		uploadRepo: uploadRepo,
		fileRepo:   fileRepo,
		storage:    storage,
	}
	return h.Request
}

func (h *requestUploadURLHandler) Request(ctx context.Context, cmd RequestUploadURLCommand) (*RequestUploadURLResult, error) {
	if err := cmd.validate(); err != nil {
		return nil, err
	}

	var ownerID *identitydomain.UserID
	if cmd.OwnerID != nil {
		uid, err := identitydomain.ParseUserID(*cmd.OwnerID)
		if err != nil {
			return nil, fmt.Errorf("invalid owner ID: %w", err)
		}
		ownerID = &uid
	}

	file, err := domain.NewFile(
		ownerID,
		cmd.Filename,
		cmd.MimeType,
		cmd.Size,
		cmd.StorageProvider,
		cmd.AccessLevel,
	)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}

	storagePath := file.GeneratePath()
	if err := file.SetPath(storagePath); err != nil {
		return nil, fmt.Errorf("set path: %w", err)
	}

	ttl := time.Duration(cmd.TTL) * time.Second
	upload, err := domain.NewFileUpload(file, ttl)
	if err != nil {
		return nil, fmt.Errorf("create upload: %w", err)
	}

	presignedURL := fmt.Sprintf("/upload/%s", file.ID())
	fields := map[string]string{
		"file_id": file.ID().String(),
		"path":    storagePath,
	}
	upload.SetUploadURL(presignedURL, fields)

	if err := h.uploadRepo.Create(ctx, upload); err != nil {
		return nil, fmt.Errorf("save upload: %w", err)
	}

	return &RequestUploadURLResult{
		UploadID:  upload.ID().String(),
		UploadURL: upload.UploadURL(),
		Fields:    upload.Fields(),
		ExpiresAt: upload.ExpiresAt().Format(time.RFC3339),
	}, nil
}

// ConfirmUploadHandler implementation
type confirmUploadHandler struct {
	uploadRepo domain.UploadRepository
	fileRepo   domain.FileRepository
	eventBus   eventbus.Bus
}

func NewConfirmUploadHandler(
	uploadRepo domain.UploadRepository,
	fileRepo domain.FileRepository,
	eventBus eventbus.Bus,
) ConfirmUploadHandler {
	h := &confirmUploadHandler{
		uploadRepo: uploadRepo,
		fileRepo:   fileRepo,
		eventBus:   eventBus,
	}
	return h.Confirm
}

func (h *confirmUploadHandler) Confirm(ctx context.Context, cmd ConfirmUploadCommand) (*ConfirmUploadResult, error) {
	uploadID, err := domain.ParseUploadID(cmd.UploadID)
	if err != nil {
		return nil, err
	}

	upload, err := h.uploadRepo.GetByID(ctx, uploadID)
	if err != nil {
		return nil, fmt.Errorf("get upload: %w", err)
	}

	if upload.IsExpired() {
		return nil, domain.ErrUploadExpired
	}

	if err := upload.MarkCompleted(); err != nil {
		return nil, err
	}

	if err := h.uploadRepo.UpdateStatus(ctx, uploadID, domain.UploadCompleted); err != nil {
		return nil, fmt.Errorf("update upload status: %w", err)
	}

	file := upload.File()
	if cmd.Checksum != "" {
		if err := file.SetChecksumFromHex(cmd.Checksum); err != nil {
			return nil, fmt.Errorf("set checksum: %w", err)
		}
	}

	if err := h.fileRepo.Create(ctx, file); err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	var ownerID *string
	if file.OwnerID() != nil {
		s := file.OwnerID().String()
		ownerID = &s
	}

	event := domain.FileUploaded{
		FileID:      file.ID(),
		OwnerID:     ownerID,
		Filename:    file.Filename(),
		MimeType:    file.MimeType(),
		Size:        file.Size(),
		StoragePath: file.Path(),
		AccessLevel: file.AccessLevel(),
		UploadedAt:  file.UploadedAt(),
	}
	_ = h.eventBus.Publish(ctx, event)

	return &ConfirmUploadResult{
		FileID:      file.ID().String(),
		Filename:    file.Filename(),
		StoredName:  file.StoredName(),
		StoragePath: file.Path(),
		UploadedAt:  file.UploadedAt().Format(time.RFC3339),
	}, nil
}

// CleanupExpiredFilesHandler implementation
type cleanupExpiredFilesHandler struct {
	fileRepo   domain.FileRepository
	uploadRepo domain.UploadRepository
	storage    StorageProvider
}

func NewCleanupExpiredFilesHandler(
	fileRepo domain.FileRepository,
	uploadRepo domain.UploadRepository,
	storage StorageProvider,
) CleanupExpiredFilesHandler {
	h := &cleanupExpiredFilesHandler{
		fileRepo:   fileRepo,
		uploadRepo: uploadRepo,
		storage:    storage,
	}
	return h.Cleanup
}

func (h *cleanupExpiredFilesHandler) Cleanup(ctx context.Context, cmd CleanupExpiredFilesCommand) (int64, error) {
	var totalDeleted int64

	expiredFiles, err := h.fileRepo.GetExpired(ctx, time.Now(), cmd.BatchSize)
	if err != nil {
		return 0, fmt.Errorf("get expired files: %w", err)
	}

	for _, file := range expiredFiles {
		_ = h.storage.Delete(ctx, file.Path())

		if err := h.fileRepo.Delete(ctx, file.ID()); err != nil {
			fmt.Printf("failed to delete file %s from repository: %v\n", file.ID(), err)
		} else {
			totalDeleted++
		}
	}

	expiredUploads, err := h.uploadRepo.GetExpired(ctx, time.Now(), cmd.BatchSize)
	if err != nil {
		return totalDeleted, nil
	}

	for _, upload := range expiredUploads {
		_ = h.uploadRepo.Delete(ctx, upload.ID())
	}

	return totalDeleted, nil
}

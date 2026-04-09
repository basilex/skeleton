package command

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/basilex/skeleton/internal/files/domain"
	identitydomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/eventbus/memory"
	"github.com/stretchr/testify/require"
)

// Mock FileRepository
type mockFileRepo struct {
	files   map[domain.FileID]*domain.File
	saved   *domain.File
	deleted bool
}

func newMockFileRepo() *mockFileRepo {
	return &mockFileRepo{
		files: make(map[domain.FileID]*domain.File),
	}
}

func (m *mockFileRepo) Create(ctx context.Context, file *domain.File) error {
	m.saved = file
	m.files[file.ID()] = file
	return nil
}

func (m *mockFileRepo) Update(ctx context.Context, file *domain.File) error {
	m.files[file.ID()] = file
	return nil
}

func (m *mockFileRepo) GetByID(ctx context.Context, id domain.FileID) (*domain.File, error) {
	if file, ok := m.files[id]; ok {
		return file, nil
	}
	return nil, domain.ErrFileNotFound
}

func (m *mockFileRepo) GetByPath(ctx context.Context, path string) (*domain.File, error) {
	for _, file := range m.files {
		if file.Path() == path {
			return file, nil
		}
	}
	return nil, domain.ErrFileNotFound
}

func (m *mockFileRepo) GetByOwner(ctx context.Context, ownerID string, limit, offset int) ([]*domain.File, error) {
	var result []*domain.File
	for _, file := range m.files {
		if file.OwnerID() != nil && file.OwnerID().String() == ownerID {
			result = append(result, file)
		}
	}
	return result, nil
}

func (m *mockFileRepo) GetExpired(ctx context.Context, before time.Time, limit int) ([]*domain.File, error) {
	var result []*domain.File
	for _, file := range m.files {
		if file.IsExpired() {
			result = append(result, file)
		}
	}
	return result, nil
}

func (m *mockFileRepo) Delete(ctx context.Context, id domain.FileID) error {
	delete(m.files, id)
	m.deleted = true
	return nil
}

func (m *mockFileRepo) DeleteBatch(ctx context.Context, ids []domain.FileID) error {
	for _, id := range ids {
		delete(m.files, id)
	}
	return nil
}

func (m *mockFileRepo) Count(ctx context.Context, filter *domain.FileFilter) (int64, error) {
	return int64(len(m.files)), nil
}

func (m *mockFileRepo) List(ctx context.Context, filter *domain.FileFilter, limit, offset int) ([]*domain.File, error) {
	var result []*domain.File
	for _, file := range m.files {
		result = append(result, file)
	}
	return result, nil
}

// Mock UploadRepository
type mockUploadRepo struct {
	uploads map[domain.UploadID]*domain.FileUpload
}

func newMockUploadRepo() *mockUploadRepo {
	return &mockUploadRepo{
		uploads: make(map[domain.UploadID]*domain.FileUpload),
	}
}

func (m *mockUploadRepo) Create(ctx context.Context, upload *domain.FileUpload) error {
	m.uploads[upload.ID()] = upload
	return nil
}

func (m *mockUploadRepo) GetByID(ctx context.Context, id domain.UploadID) (*domain.FileUpload, error) {
	if upload, ok := m.uploads[id]; ok {
		return upload, nil
	}
	return nil, domain.ErrUploadNotFound
}

func (m *mockUploadRepo) GetByFileID(ctx context.Context, fileID domain.FileID) (*domain.FileUpload, error) {
	for _, upload := range m.uploads {
		if upload.File().ID() == fileID {
			return upload, nil
		}
	}
	return nil, domain.ErrUploadNotFound
}

func (m *mockUploadRepo) UpdateStatus(ctx context.Context, id domain.UploadID, status domain.UploadStatus) error {
	if upload, ok := m.uploads[id]; ok {
		_ = upload.MarkCompleted()
	}
	return nil
}

func (m *mockUploadRepo) GetExpired(ctx context.Context, before time.Time, limit int) ([]*domain.FileUpload, error) {
	var result []*domain.FileUpload
	for _, upload := range m.uploads {
		if upload.IsExpired() {
			result = append(result, upload)
		}
	}
	return result, nil
}

func (m *mockUploadRepo) Delete(ctx context.Context, id domain.UploadID) error {
	delete(m.uploads, id)
	return nil
}

func (m *mockUploadRepo) DeleteByFileID(ctx context.Context, fileID domain.FileID) error {
	for id, upload := range m.uploads {
		if upload.File().ID() == fileID {
			delete(m.uploads, id)
		}
	}
	return nil
}

// Mock ProcessingRepository
type mockProcessingRepo struct {
	processings map[domain.ProcessingID]*domain.FileProcessing
}

func newMockProcessingRepo() *mockProcessingRepo {
	return &mockProcessingRepo{
		processings: make(map[domain.ProcessingID]*domain.FileProcessing),
	}
}

func (m *mockProcessingRepo) Create(ctx context.Context, processing *domain.FileProcessing) error {
	m.processings[processing.ID()] = processing
	return nil
}

func (m *mockProcessingRepo) Update(ctx context.Context, processing *domain.FileProcessing) error {
	m.processings[processing.ID()] = processing
	return nil
}

func (m *mockProcessingRepo) GetByID(ctx context.Context, id domain.ProcessingID) (*domain.FileProcessing, error) {
	if processing, ok := m.processings[id]; ok {
		return processing, nil
	}
	return nil, domain.ErrProcessingNotFound
}

func (m *mockProcessingRepo) GetByFileID(ctx context.Context, fileID domain.FileID) ([]*domain.FileProcessing, error) {
	var result []*domain.FileProcessing
	for _, p := range m.processings {
		if p.FileID() == fileID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockProcessingRepo) GetPending(ctx context.Context, limit int) ([]*domain.FileProcessing, error) {
	var result []*domain.FileProcessing
	for _, p := range m.processings {
		if p.Status() == domain.ProcessingPending {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockProcessingRepo) GetByStatus(ctx context.Context, status domain.ProcessingStatus, limit, offset int) ([]*domain.FileProcessing, error) {
	var result []*domain.FileProcessing
	for _, p := range m.processings {
		if p.Status() == status {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *mockProcessingRepo) Delete(ctx context.Context, id domain.ProcessingID) error {
	delete(m.processings, id)
	return nil
}

func (m *mockProcessingRepo) DeleteByFileID(ctx context.Context, fileID domain.FileID) error {
	for id, p := range m.processings {
		if p.FileID() == fileID {
			delete(m.processings, id)
		}
	}
	return nil
}

func (m *mockProcessingRepo) Count(ctx context.Context, filter *domain.ProcessingFilter) (int64, error) {
	return int64(len(m.processings)), nil
}

func (m *mockProcessingRepo) List(ctx context.Context, filter *domain.ProcessingFilter, limit, offset int) ([]*domain.FileProcessing, error) {
	var result []*domain.FileProcessing
	for _, p := range m.processings {
		if filter.FileID != nil && p.FileID() != *filter.FileID {
			continue
		}
		result = append(result, p)
	}
	return result, nil
}

// Mock StorageProvider
type mockStorageProvider struct {
	uploadedPath string
	uploadError  error
	deleteError  error
}

func (m *mockStorageProvider) Upload(ctx context.Context, path string, reader io.Reader, contentType string) error {
	if m.uploadError != nil {
		return m.uploadError
	}
	m.uploadedPath = path
	return nil
}

func (m *mockStorageProvider) Delete(ctx context.Context, path string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	return nil
}

// Tests
func TestUploadFileHandler(t *testing.T) {
	t.Run("successful upload with owner", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		storage := &mockStorageProvider{}
		bus := memory.New()
		handler := NewUploadFileHandler(fileRepo, storage, bus)

		user := identitydomain.NewUserID()
		userID := user.String()
		content := bytes.NewReader([]byte("test content"))

		result, err := handler(context.Background(), UploadFileCommand{
			OwnerID:         &userID,
			Filename:        "test.jpg",
			MimeType:        "image/jpeg",
			Size:            1024,
			Content:         content,
			StorageProvider: domain.StorageLocal,
			AccessLevel:     domain.AccessPrivate,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotEmpty(t, result.FileID)
		require.Equal(t, "test.jpg", fileRepo.saved.Filename())
		require.Equal(t, domain.StorageLocal, fileRepo.saved.StorageProvider())
		require.NotEmpty(t, storage.uploadedPath)
	})

	t.Run("anonymous upload", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		storage := &mockStorageProvider{}
		bus := memory.New()
		handler := NewUploadFileHandler(fileRepo, storage, bus)

		content := bytes.NewReader([]byte("public file"))

		result, err := handler(context.Background(), UploadFileCommand{
			OwnerID:         nil,
			Filename:        "public.pdf",
			MimeType:        "application/pdf",
			Size:            2048,
			Content:         content,
			StorageProvider: domain.StorageS3,
			AccessLevel:     domain.AccessPublic,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Nil(t, fileRepo.saved.OwnerID())
		require.Equal(t, domain.AccessPublic, fileRepo.saved.AccessLevel())
	})

	t.Run("validation error - empty filename", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		storage := &mockStorageProvider{}
		bus := memory.New()
		handler := NewUploadFileHandler(fileRepo, storage, bus)

		content := bytes.NewReader([]byte("test"))

		_, err := handler(context.Background(), UploadFileCommand{
			Filename:        "",
			MimeType:        "image/jpeg",
			Size:            1024,
			Content:         content,
			StorageProvider: domain.StorageLocal,
			AccessLevel:     domain.AccessPrivate,
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "filename")
	})

	t.Run("storage upload failure", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		storage := &mockStorageProvider{
			uploadError: errors.New("storage error"),
		}
		bus := memory.New()
		handler := NewUploadFileHandler(fileRepo, storage, bus)

		content := bytes.NewReader([]byte("test"))

		_, err := handler(context.Background(), UploadFileCommand{
			Filename:        "test.jpg",
			MimeType:        "image/jpeg",
			Size:            1024,
			Content:         content,
			StorageProvider: domain.StorageLocal,
			AccessLevel:     domain.AccessPrivate,
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "upload to storage")
	})
}

func TestDeleteFileHandler(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		storage := &mockStorageProvider{}
		bus := memory.New()

		userID := identitydomain.NewUserID()
		file, _ := domain.NewFile(&userID, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPrivate)
		_ = fileRepo.Create(context.Background(), file)

		handler := NewDeleteFileHandler(fileRepo, storage, bus)
		err := handler(context.Background(), DeleteFileCommand{
			FileID: file.ID().String(),
		})

		require.NoError(t, err)
		require.True(t, fileRepo.deleted)
	})

	t.Run("file not found", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		storage := &mockStorageProvider{}
		bus := memory.New()
		handler := NewDeleteFileHandler(fileRepo, storage, bus)

		err := handler(context.Background(), DeleteFileCommand{
			FileID: domain.NewFileID().String(),
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "get file")
	})
}

func TestRequestUploadURLHandler(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		uploadRepo := newMockUploadRepo()
		fileRepo := newMockFileRepo()
		storage := &mockStorageProvider{}

		handler := NewRequestUploadURLHandler(uploadRepo, fileRepo, storage)

		user := identitydomain.NewUserID()
		userID := user.String()
		result, err := handler(context.Background(), RequestUploadURLCommand{
			OwnerID:         &userID,
			Filename:        "large-file.zip",
			MimeType:        "application/zip",
			Size:            10 * 1024 * 1024,
			StorageProvider: domain.StorageS3,
			AccessLevel:     domain.AccessPrivate,
			TTL:             3600,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.NotEmpty(t, result.UploadID)
		require.NotEmpty(t, result.UploadURL)
		require.NotEmpty(t, result.Fields)
	})

	t.Run("validation error - invalid provider", func(t *testing.T) {
		uploadRepo := newMockUploadRepo()
		fileRepo := newMockFileRepo()
		storage := &mockStorageProvider{}

		handler := NewRequestUploadURLHandler(uploadRepo, fileRepo, storage)

		_, err := handler(context.Background(), RequestUploadURLCommand{
			Filename:        "test.jpg",
			MimeType:        "image/jpeg",
			Size:            1024,
			StorageProvider: domain.StorageProvider("invalid"),
			AccessLevel:     domain.AccessPrivate,
			TTL:             3600,
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "storage provider")
	})
}

func TestConfirmUploadHandler(t *testing.T) {
	t.Run("successful confirmation", func(t *testing.T) {
		uploadRepo := newMockUploadRepo()
		fileRepo := newMockFileRepo()
		bus := memory.New()

		userID := identitydomain.NewUserID()
		file, _ := domain.NewFile(&userID, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPrivate)
		upload, _ := domain.NewFileUpload(file, time.Hour)
		_ = uploadRepo.Create(context.Background(), upload)

		handler := NewConfirmUploadHandler(uploadRepo, fileRepo, bus)
		result, err := handler(context.Background(), ConfirmUploadCommand{
			UploadID: upload.ID().String(),
			Checksum: "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3d391e987982fbbd3d391e987",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "test.jpg", result.Filename)
	})

	t.Run("upload not found", func(t *testing.T) {
		uploadRepo := newMockUploadRepo()
		fileRepo := newMockFileRepo()
		bus := memory.New()
		handler := NewConfirmUploadHandler(uploadRepo, fileRepo, bus)

		_, err := handler(context.Background(), ConfirmUploadCommand{
			UploadID: domain.NewUploadID().String(),
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "get upload")
	})
}

func TestCleanupExpiredFilesHandler(t *testing.T) {
	t.Run("successful cleanup", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		uploadRepo := newMockUploadRepo()
		storage := &mockStorageProvider{}

		userID := identitydomain.NewUserID()
		pastTime := time.Now().Add(-1 * time.Hour)

		file := domain.ReconstituteFile(
			domain.NewFileID(),
			&userID,
			"expired.jpg",
			"expired.jpg",
			"image/jpeg",
			1024,
			"uploads/expired.jpg",
			domain.StorageLocal,
			"abc123",
			domain.FileMetadata{},
			domain.AccessPrivate,
			domain.ScanStatusPending,
			"",
			nil,
			time.Now().Add(-2*time.Hour),
			&pastTime,
			nil,
			time.Now().Add(-2*time.Hour),
			time.Now().Add(-2*time.Hour),
		)
		_ = fileRepo.Create(context.Background(), file)

		handler := NewCleanupExpiredFilesHandler(fileRepo, uploadRepo, storage)
		count, err := handler(context.Background(), CleanupExpiredFilesCommand{
			BatchSize: 10,
		})

		require.NoError(t, err)
		require.Equal(t, int64(1), count)
	})
}

package query

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/skeleton/internal/files/domain"
	identitydomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/stretchr/testify/require"
)

// Mock repositories (reusing from command tests)
type mockFileRepo struct {
	files map[domain.FileID]*domain.File
}

func newMockFileRepo() *mockFileRepo {
	return &mockFileRepo{
		files: make(map[domain.FileID]*domain.File),
	}
}

func (m *mockFileRepo) Create(ctx context.Context, file *domain.File) error {
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
		if file.OwnerID() != nil && string(*file.OwnerID()) == ownerID {
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
		if filter != nil {
			if filter.OwnerID != nil && (file.OwnerID() == nil || string(*file.OwnerID()) != *filter.OwnerID) {
				continue
			}
			if filter.MimeType != nil && file.MimeType() != *filter.MimeType {
				continue
			}
			if filter.AccessLevel != nil && file.AccessLevel() != *filter.AccessLevel {
				continue
			}
		}
		result = append(result, file)
	}
	return result, nil
}

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
		if filter != nil {
			if filter.FileID != nil && p.FileID() != *filter.FileID {
				continue
			}
			if filter.Status != nil && p.Status() != *filter.Status {
				continue
			}
		}
		result = append(result, p)
	}
	return result, nil
}

// Tests
func TestGetFileHandler(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		handler := NewGetFileHandler(fileRepo)

		userID := identitydomain.NewUserID()
		file, _ := domain.NewFile(&userID, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPrivate)
		_ = fileRepo.Create(context.Background(), file)

		result, err := handler(context.Background(), GetFileQuery{
			FileID: file.ID().String(),
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, file.ID().String(), result.ID)
		require.Equal(t, "test.jpg", result.Filename)
		require.Equal(t, "image/jpeg", result.MimeType)
		require.Equal(t, int64(1024), result.Size)
	})

	t.Run("file not found", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		handler := NewGetFileHandler(fileRepo)

		_, err := handler(context.Background(), GetFileQuery{
			FileID: domain.NewFileID().String(),
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "get file")
	})

	t.Run("invalid file ID", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		handler := NewGetFileHandler(fileRepo)

		_, err := handler(context.Background(), GetFileQuery{
			FileID: "invalid-id",
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "file")
	})
}

func TestListFilesHandler(t *testing.T) {
	t.Run("list all files", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		handler := NewListFilesHandler(fileRepo)

		userID := identitydomain.NewUserID()
		file1, _ := domain.NewFile(&userID, "file1.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPrivate)
		file2, _ := domain.NewFile(&userID, "file2.pdf", "application/pdf", 2048, domain.StorageS3, domain.AccessPublic)
		_ = fileRepo.Create(context.Background(), file1)
		_ = fileRepo.Create(context.Background(), file2)

		result, err := handler(context.Background(), ListFilesQuery{
			Limit: 10,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 2)
		require.False(t, result.HasMore)
	})

	t.Run("list with owner filter", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		handler := NewListFilesHandler(fileRepo)

		userID1 := identitydomain.NewUserID()
		userID2 := identitydomain.NewUserID()
		file1, _ := domain.NewFile(&userID1, "file1.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPrivate)
		file2, _ := domain.NewFile(&userID2, "file2.pdf", "application/pdf", 2048, domain.StorageS3, domain.AccessPublic)
		_ = fileRepo.Create(context.Background(), file1)
		_ = fileRepo.Create(context.Background(), file2)

		ownerID := string(userID1)
		result, err := handler(context.Background(), ListFilesQuery{
			OwnerID: &ownerID,
			Limit:   10,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 1)
		require.NotNil(t, result.Items[0].OwnerID)
		require.Equal(t, string(userID1), *result.Items[0].OwnerID)
	})

	t.Run("list with pagination", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		handler := NewListFilesHandler(fileRepo)

		userID := identitydomain.NewUserID()
		for i := 0; i < 25; i++ {
			file, _ := domain.NewFile(&userID, "file.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPrivate)
			_ = fileRepo.Create(context.Background(), file)
		}

		result, err := handler(context.Background(), ListFilesQuery{
			Limit: 10,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.HasMore)
		require.NotNil(t, result.NextCursor)
	})

	t.Run("default limit", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		handler := NewListFilesHandler(fileRepo)

		result, err := handler(context.Background(), ListFilesQuery{
			Limit: 0,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
	})

	t.Run("max limit", func(t *testing.T) {
		fileRepo := newMockFileRepo()
		handler := NewListFilesHandler(fileRepo)

		result, err := handler(context.Background(), ListFilesQuery{
			Limit: 200,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
	})
}

func TestGetProcessingStatusHandler(t *testing.T) {
	t.Run("successful get status", func(t *testing.T) {
		processingRepo := newMockProcessingRepo()
		handler := NewGetProcessingStatusHandler(processingRepo)

		fileID := domain.NewFileID()
		processing, _ := domain.NewFileProcessing(fileID, domain.OperationResize, domain.ProcessingOptions{})
		_ = processingRepo.Create(context.Background(), processing)

		result, err := handler(context.Background(), GetProcessingStatusQuery{
			ProcessingID: processing.ID().String(),
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, processing.ID().String(), result.ID)
		require.Equal(t, fileID.String(), result.FileID)
		require.Equal(t, string(domain.ProcessingPending), result.Status)
	})

	t.Run("processing not found", func(t *testing.T) {
		processingRepo := newMockProcessingRepo()
		handler := NewGetProcessingStatusHandler(processingRepo)

		_, err := handler(context.Background(), GetProcessingStatusQuery{
			ProcessingID: domain.NewProcessingID().String(),
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "get processing")
	})

	t.Run("invalid processing ID", func(t *testing.T) {
		processingRepo := newMockProcessingRepo()
		handler := NewGetProcessingStatusHandler(processingRepo)

		_, err := handler(context.Background(), GetProcessingStatusQuery{
			ProcessingID: "invalid-id",
		})

		require.Error(t, err)
		require.Contains(t, err.Error(), "processing")
	})
}

func TestListProcessingsHandler(t *testing.T) {
	t.Run("list processings for file", func(t *testing.T) {
		processingRepo := newMockProcessingRepo()
		handler := NewListProcessingsHandler(processingRepo)

		fileID1 := domain.NewFileID()
		fileID2 := domain.NewFileID()

		processing1, _ := domain.NewFileProcessing(fileID1, domain.OperationResize, domain.ProcessingOptions{})
		processing2, _ := domain.NewFileProcessing(fileID1, domain.OperationCrop, domain.ProcessingOptions{})
		processing3, _ := domain.NewFileProcessing(fileID2, domain.OperationCompress, domain.ProcessingOptions{})

		_ = processingRepo.Create(context.Background(), processing1)
		_ = processingRepo.Create(context.Background(), processing2)
		_ = processingRepo.Create(context.Background(), processing3)

		result, err := handler(context.Background(), ListProcessingsQuery{
			FileID: fileID1.String(),
			Limit:  10,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 2)
	})

	t.Run("list with status filter", func(t *testing.T) {
		processingRepo := newMockProcessingRepo()
		handler := NewListProcessingsHandler(processingRepo)

		fileID := domain.NewFileID()

		processing1, _ := domain.NewFileProcessing(fileID, domain.OperationResize, domain.ProcessingOptions{})
		processing2, _ := domain.NewFileProcessing(fileID, domain.OperationCrop, domain.ProcessingOptions{})
		_ = processing2.Start() // running status

		_ = processingRepo.Create(context.Background(), processing1)
		_ = processingRepo.Create(context.Background(), processing2)

		status := domain.ProcessingRunning
		result, err := handler(context.Background(), ListProcessingsQuery{
			FileID: fileID.String(),
			Status: &status,
			Limit:  10,
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 1)
		require.Equal(t, string(domain.ProcessingRunning), result.Items[0].Status)
	})

	t.Run("invalid file ID", func(t *testing.T) {
		processingRepo := newMockProcessingRepo()
		handler := NewListProcessingsHandler(processingRepo)

		// ListProcessings doesn't fail on invalid ID, it just returns empty list
		result, err := handler(context.Background(), ListProcessingsQuery{
			FileID: "invalid-id",
			Limit:  10,
		})

		// Should not error, just return empty list
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Items, 0)
	})
}

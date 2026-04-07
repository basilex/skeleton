package persistence

import (
	"context"
	"testing"

	"github.com/basilex/skeleton/internal/files/domain"
	"github.com/stretchr/testify/require"
)

func TestProcessingRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	processingRepo := NewProcessingRepository(db)
	fileRepo := NewFileRepository(db)
	ctx := context.Background()

	t.Run("create processing", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = fileRepo.Create(ctx, file)

		processing, _ := domain.NewFileProcessing(file.ID(), domain.OperationResize, domain.ProcessingOptions{})
		err := processingRepo.Create(ctx, processing)
		require.NoError(t, err)

		retrieved, err := processingRepo.GetByID(ctx, processing.ID())
		require.NoError(t, err)
		require.Equal(t, processing.ID(), retrieved.ID())
		require.Equal(t, file.ID(), retrieved.FileID())
		require.Equal(t, domain.ProcessingPending, retrieved.Status())
	})
}

func TestProcessingRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	processingRepo := NewProcessingRepository(db)
	fileRepo := NewFileRepository(db)
	ctx := context.Background()

	t.Run("get existing processing", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = fileRepo.Create(ctx, file)

		processing, _ := domain.NewFileProcessing(file.ID(), domain.OperationResize, domain.ProcessingOptions{})
		_ = processingRepo.Create(ctx, processing)

		retrieved, err := processingRepo.GetByID(ctx, processing.ID())
		require.NoError(t, err)
		require.Equal(t, processing.ID(), retrieved.ID())
	})

	t.Run("get non-existent processing", func(t *testing.T) {
		_, err := processingRepo.GetByID(ctx, domain.NewProcessingID())
		require.Error(t, err)
		require.Equal(t, domain.ErrProcessingNotFound, err)
	})
}

func TestProcessingRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	processingRepo := NewProcessingRepository(db)
	fileRepo := NewFileRepository(db)
	ctx := context.Background()

	t.Run("update processing status", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = fileRepo.Create(ctx, file)

		processing, _ := domain.NewFileProcessing(file.ID(), domain.OperationResize, domain.ProcessingOptions{})
		_ = processingRepo.Create(ctx, processing)

		_ = processing.Start()
		_ = processingRepo.Update(ctx, processing)

		retrieved, _ := processingRepo.GetByID(ctx, processing.ID())
		require.Equal(t, domain.ProcessingRunning, retrieved.Status())
	})
}

func TestProcessingRepository_GetByFileID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	processingRepo := NewProcessingRepository(db)
	fileRepo := NewFileRepository(db)
	ctx := context.Background()

	t.Run("get processings by file ID", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = fileRepo.Create(ctx, file)

		processing1, _ := domain.NewFileProcessing(file.ID(), domain.OperationResize, domain.ProcessingOptions{})
		processing2, _ := domain.NewFileProcessing(file.ID(), domain.OperationCrop, domain.ProcessingOptions{})

		_ = processingRepo.Create(ctx, processing1)
		_ = processingRepo.Create(ctx, processing2)

		processings, err := processingRepo.GetByFileID(ctx, file.ID())
		require.NoError(t, err)
		require.Len(t, processings, 2)
	})
}

func TestProcessingRepository_GetPending(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	processingRepo := NewProcessingRepository(db)
	fileRepo := NewFileRepository(db)
	ctx := context.Background()

	t.Run("get pending processings", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = fileRepo.Create(ctx, file)

		// Create pending processing
		pending1, _ := domain.NewFileProcessing(file.ID(), domain.OperationResize, domain.ProcessingOptions{})
		_ = processingRepo.Create(ctx, pending1)

		// Create running processing
		running, _ := domain.NewFileProcessing(file.ID(), domain.OperationCrop, domain.ProcessingOptions{})
		_ = running.Start()
		_ = processingRepo.Create(ctx, running)

		// Create completed processing
		completed, _ := domain.NewFileProcessing(file.ID(), domain.OperationCompress, domain.ProcessingOptions{})
		_ = completed.Start()
		resultFile, _ := domain.NewFile(nil, "result.jpg", "image/jpeg", 512, domain.StorageLocal, domain.AccessPublic)
		_ = completed.Complete(resultFile.ID())
		_ = processingRepo.Create(ctx, completed)

		pendings, err := processingRepo.GetPending(ctx, 10)
		require.NoError(t, err)
		require.Len(t, pendings, 1)
		require.Equal(t, pending1.ID(), pendings[0].ID())
	})
}

func TestProcessingRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	processingRepo := NewProcessingRepository(db)
	fileRepo := NewFileRepository(db)
	ctx := context.Background()

	t.Run("delete processing", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = fileRepo.Create(ctx, file)

		processing, _ := domain.NewFileProcessing(file.ID(), domain.OperationResize, domain.ProcessingOptions{})
		_ = processingRepo.Create(ctx, processing)

		err := processingRepo.Delete(ctx, processing.ID())
		require.NoError(t, err)

		_, err = processingRepo.GetByID(ctx, processing.ID())
		require.Error(t, err)
	})
}

func TestProcessingRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	processingRepo := NewProcessingRepository(db)
	fileRepo := NewFileRepository(db)
	ctx := context.Background()

	t.Run("list with filter", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = fileRepo.Create(ctx, file)

		for i := 0; i < 3; i++ {
			processing, _ := domain.NewFileProcessing(file.ID(), domain.OperationResize, domain.ProcessingOptions{})
			_ = processingRepo.Create(ctx, processing)
		}

		fileID := file.ID()
		processings, err := processingRepo.List(ctx, &domain.ProcessingFilter{FileID: &fileID}, 10, 0)
		require.NoError(t, err)
		require.Len(t, processings, 3)
	})
}

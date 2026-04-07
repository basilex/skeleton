package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/skeleton/internal/files/domain"
	"github.com/stretchr/testify/require"
)

func TestUploadRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	uploadRepo := NewUploadRepository(db)
	fileRepo := NewFileRepository(db)
	ctx := context.Background()

	t.Run("create upload", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = fileRepo.Create(ctx, file)

		upload, _ := domain.NewFileUpload(file, time.Hour)
		err := uploadRepo.Create(ctx, upload)
		require.NoError(t, err)

		retrieved, err := uploadRepo.GetByID(ctx, upload.ID())
		require.NoError(t, err)
		require.Equal(t, upload.ID(), retrieved.ID())
		require.Equal(t, file.ID(), retrieved.File().ID())
	})
}

func TestUploadRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	uploadRepo := NewUploadRepository(db)
	fileRepo := NewFileRepository(db)
	ctx := context.Background()

	t.Run("get existing upload", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = fileRepo.Create(ctx, file)

		upload, _ := domain.NewFileUpload(file, time.Hour)
		_ = uploadRepo.Create(ctx, upload)

		retrieved, err := uploadRepo.GetByID(ctx, upload.ID())
		require.NoError(t, err)
		require.Equal(t, upload.ID(), retrieved.ID())
	})

	t.Run("get non-existent upload", func(t *testing.T) {
		_, err := uploadRepo.GetByID(ctx, domain.NewUploadID())
		require.Error(t, err)
		require.Equal(t, domain.ErrUploadNotFound, err)
	})
}

func TestUploadRepository_GetByFileID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	uploadRepo := NewUploadRepository(db)
	fileRepo := NewFileRepository(db)
	ctx := context.Background()

	t.Run("get upload by file ID", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = fileRepo.Create(ctx, file)

		upload, _ := domain.NewFileUpload(file, time.Hour)
		_ = uploadRepo.Create(ctx, upload)

		retrieved, err := uploadRepo.GetByFileID(ctx, file.ID())
		require.NoError(t, err)
		require.Equal(t, upload.ID(), retrieved.ID())
	})
}

func TestUploadRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	uploadRepo := NewUploadRepository(db)
	fileRepo := NewFileRepository(db)
	ctx := context.Background()

	t.Run("update status", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = fileRepo.Create(ctx, file)

		upload, _ := domain.NewFileUpload(file, time.Hour)
		_ = uploadRepo.Create(ctx, upload)

		err := uploadRepo.UpdateStatus(ctx, upload.ID(), domain.UploadCompleted)
		require.NoError(t, err)

		retrieved, _ := uploadRepo.GetByID(ctx, upload.ID())
		require.Equal(t, domain.UploadCompleted, retrieved.Status())
	})
}

func TestUploadRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	uploadRepo := NewUploadRepository(db)
	fileRepo := NewFileRepository(db)
	ctx := context.Background()

	t.Run("delete upload", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = fileRepo.Create(ctx, file)

		upload, _ := domain.NewFileUpload(file, time.Hour)
		_ = uploadRepo.Create(ctx, upload)

		err := uploadRepo.Delete(ctx, upload.ID())
		require.NoError(t, err)

		_, err = uploadRepo.GetByID(ctx, upload.ID())
		require.Error(t, err)
	})
}

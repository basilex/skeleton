package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/basilex/skeleton/internal/files/domain"
	identitydomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestFileRepository_Create(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := NewFileRepository(pool)
	ctx := context.Background()

	t.Run("create file with owner", func(t *testing.T) {
		userID := identitydomain.NewUserID()
		file, err := domain.NewFile(&userID, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPrivate)
		require.NoError(t, err)

		err = repo.Create(ctx, file)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, file.ID())
		require.NoError(t, err)
		require.Equal(t, file.ID(), retrieved.ID())
		require.Equal(t, "test.jpg", retrieved.Filename())
		require.NotNil(t, retrieved.OwnerID())
	})

	t.Run("create file without owner", func(t *testing.T) {
		file, err := domain.NewFile(nil, "public.pdf", "application/pdf", 2048, domain.StorageS3, domain.AccessPublic)
		require.NoError(t, err)

		err = repo.Create(ctx, file)
		require.NoError(t, err)

		retrieved, err := repo.GetByID(ctx, file.ID())
		require.NoError(t, err)
		require.Nil(t, retrieved.OwnerID())
	})

	t.Run("create duplicate file", func(t *testing.T) {
		file, err := domain.NewFile(nil, "duplicate.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		require.NoError(t, err)

		err = repo.Create(ctx, file)
		require.NoError(t, err)

		err = repo.Create(ctx, file)
		require.Error(t, err)
	})
}

func TestFileRepository_GetByID(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := NewFileRepository(pool)
	ctx := context.Background()

	t.Run("get existing file", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "test.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = repo.Create(ctx, file)

		retrieved, err := repo.GetByID(ctx, file.ID())
		require.NoError(t, err)
		require.Equal(t, file.ID(), retrieved.ID())
	})

	t.Run("get non-existent file", func(t *testing.T) {
		_, err := repo.GetByID(ctx, domain.NewFileID())
		require.Error(t, err)
		require.Equal(t, domain.ErrFileNotFound, err)
	})
}

func TestFileRepository_Update(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := NewFileRepository(pool)
	ctx := context.Background()

	t.Run("update file", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "original.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = repo.Create(ctx, file)

		_ = file.SetPath("new/path.jpg")
		err := repo.Update(ctx, file)
		require.NoError(t, err)

		retrieved, _ := repo.GetByID(ctx, file.ID())
		require.Equal(t, "new/path.jpg", retrieved.Path())
	})
}

func TestFileRepository_Delete(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := NewFileRepository(pool)
	ctx := context.Background()

	t.Run("delete existing file", func(t *testing.T) {
		file, _ := domain.NewFile(nil, "delete.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		_ = repo.Create(ctx, file)

		err := repo.Delete(ctx, file.ID())
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, file.ID())
		require.Error(t, err)
		require.Equal(t, domain.ErrFileNotFound, err)
	})

	t.Run("delete non-existent file", func(t *testing.T) {
		err := repo.Delete(ctx, domain.NewFileID())
		require.Error(t, err)
		require.Equal(t, domain.ErrFileNotFound, err)
	})
}

func TestFileRepository_GetByOwner(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := NewFileRepository(pool)
	ctx := context.Background()

	t.Run("get files by owner", func(t *testing.T) {
		userID := identitydomain.NewUserID()
		otherUserID := identitydomain.NewUserID()

		file1, _ := domain.NewFile(&userID, "file1.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPrivate)
		file2, _ := domain.NewFile(&userID, "file2.jpg", "image/jpeg", 2048, domain.StorageLocal, domain.AccessPrivate)
		file3, _ := domain.NewFile(&otherUserID, "file3.jpg", "image/jpeg", 4096, domain.StorageLocal, domain.AccessPrivate)

		_ = repo.Create(ctx, file1)
		_ = repo.Create(ctx, file2)
		_ = repo.Create(ctx, file3)

		files, err := repo.GetByOwner(ctx, userID.String(), 10, 0)
		require.NoError(t, err)
		require.Len(t, files, 2)
	})

	t.Run("get files with pagination", func(t *testing.T) {
		userID := identitydomain.NewUserID()

		for i := 0; i < 5; i++ {
			file, _ := domain.NewFile(&userID, "file.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPrivate)
			_ = repo.Create(ctx, file)
		}

		files, err := repo.GetByOwner(ctx, userID.String(), 3, 0)
		require.NoError(t, err)
		require.Len(t, files, 3)

		files2, err := repo.GetByOwner(ctx, userID.String(), 3, 3)
		require.NoError(t, err)
		require.Len(t, files2, 2)
	})
}

func TestFileRepository_GetExpired(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := NewFileRepository(pool)
	ctx := context.Background()

	t.Run("get expired files", func(t *testing.T) {
		userID := identitydomain.NewUserID()

		pastTime := time.Now().Add(-1 * time.Hour)
		expiredFile := domain.ReconstituteFile(
			domain.NewFileID(),
			&userID,
			"expired.jpg",
			"expired.jpg",
			"image/jpeg",
			1024,
			"expired.jpg",
			domain.StorageLocal,
			"abc123",
			domain.FileMetadata{},
			domain.AccessPrivate,
			time.Now().Add(-2*time.Hour),
			&pastTime,
			nil,
			time.Now().Add(-2*time.Hour),
			time.Now().Add(-2*time.Hour),
		)
		_ = repo.Create(ctx, expiredFile)

		file, _ := domain.NewFile(&userID, "valid.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPrivate)
		_ = repo.Create(ctx, file)

		files, err := repo.GetExpired(ctx, time.Now(), 10)
		require.NoError(t, err)
		require.Len(t, files, 1)
		require.Equal(t, expiredFile.ID(), files[0].ID())
	})
}

func TestFileRepository_List(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := NewFileRepository(pool)
	ctx := context.Background()

	t.Run("list all files", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			file, _ := domain.NewFile(nil, "file.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
			_ = repo.Create(ctx, file)
		}

		files, err := repo.List(ctx, nil, 10, 0)
		require.NoError(t, err)
		require.Len(t, files, 5)
	})

	t.Run("list with filter", func(t *testing.T) {
		userID := identitydomain.NewUserID()

		publicFile, _ := domain.NewFile(nil, "public.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
		privateFile, _ := domain.NewFile(&userID, "private.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPrivate)

		_ = repo.Create(ctx, publicFile)
		_ = repo.Create(ctx, privateFile)

		accessLevel := domain.AccessPublic
		files, err := repo.List(ctx, &domain.FileFilter{AccessLevel: &accessLevel}, 10, 0)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(files), 1)
	})
}

func TestFileRepository_Count(t *testing.T) {
	pool := testutil.SetupPostgres(t)
	testutil.RunMigrations(t, pool, testutil.DefaultSchema)

	repo := NewFileRepository(pool)
	ctx := context.Background()

	t.Run("count files", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			file, _ := domain.NewFile(nil, "file.jpg", "image/jpeg", 1024, domain.StorageLocal, domain.AccessPublic)
			_ = repo.Create(ctx, file)
		}

		count, err := repo.Count(ctx, nil)
		require.NoError(t, err)
		require.Equal(t, int64(3), count)
	})
}

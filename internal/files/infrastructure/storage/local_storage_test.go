package storage

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLocalStorage(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "test-local-storage")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)
		require.NotNil(t, storage)
		require.Equal(t, "local", storage.Name())

		// Check directory was created
		_, err = os.Stat(tmpDir)
		require.NoError(t, err)
	})

	t.Run("existing directory", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "existing-dir")
		os.MkdirAll(tmpDir, 0755)
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)
		require.NotNil(t, storage)
	})
}

func TestLocalStorage_Upload(t *testing.T) {
	t.Run("successful upload", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "upload-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		content := []byte("test file content")
		reader := bytes.NewReader(content)

		err = storage.Upload(context.Background(), "test/file.txt", reader, "text/plain")
		require.NoError(t, err)

		// Verify file was created
		fullPath := filepath.Join(tmpDir, "test/file.txt")
		_, err = os.Stat(fullPath)
		require.NoError(t, err)
	})

	t.Run("upload creates directories", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "upload-dirs-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		content := []byte("nested file")
		reader := bytes.NewReader(content)

		err = storage.Upload(context.Background(), "deeply/nested/path/file.txt", reader, "text/plain")
		require.NoError(t, err)

		// Verify nested directory was created
		fullPath := filepath.Join(tmpDir, "deeply/nested/path/file.txt")
		_, err = os.Stat(fullPath)
		require.NoError(t, err)
	})
}

func TestLocalStorage_UploadFromBytes(t *testing.T) {
	t.Run("successful upload", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "upload-bytes-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		content := []byte("test content from bytes")

		err = storage.UploadFromBytes(context.Background(), "test.txt", content, "text/plain")
		require.NoError(t, err)

		// Verify content
		fullPath := filepath.Join(tmpDir, "test.txt")
		data, err := os.ReadFile(fullPath)
		require.NoError(t, err)
		require.Equal(t, content, data)
	})
}

func TestLocalStorage_Download(t *testing.T) {
	t.Run("successful download", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "download-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		// Upload first
		content := []byte("download test content")
		err = storage.UploadFromBytes(context.Background(), "download.txt", content, "text/plain")
		require.NoError(t, err)

		// Download
		reader, err := storage.Download(context.Background(), "download.txt")
		require.NoError(t, err)
		defer reader.Close()

		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(reader)
		require.NoError(t, err)
		require.Equal(t, content, buf.Bytes())
	})

	t.Run("file not found", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "download-notfound-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		_, err = storage.Download(context.Background(), "nonexistent.txt")
		require.Error(t, err)
		require.Equal(t, ErrFileNotFound, err)
	})
}

func TestLocalStorage_DownloadToBytes(t *testing.T) {
	t.Run("successful download", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "download-bytes-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		content := []byte("test download to bytes")

		err = storage.UploadFromBytes(context.Background(), "test.txt", content, "text/plain")
		require.NoError(t, err)

		data, err := storage.DownloadToBytes(context.Background(), "test.txt")
		require.NoError(t, err)
		require.Equal(t, content, data)
	})
}

func TestLocalStorage_Delete(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "delete-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		// Upload first
		err = storage.UploadFromBytes(context.Background(), "delete.txt", []byte("content"), "text/plain")
		require.NoError(t, err)

		// Delete
		err = storage.Delete(context.Background(), "delete.txt")
		require.NoError(t, err)

		// Verify deleted
		fullPath := filepath.Join(tmpDir, "delete.txt")
		_, err = os.Stat(fullPath)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("delete non-existent file", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "delete-nonexistent-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		// Should not error
		err = storage.Delete(context.Background(), "nonexistent.txt")
		require.NoError(t, err)
	})
}

func TestLocalStorage_Exists(t *testing.T) {
	t.Run("file exists", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "exists-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		err = storage.UploadFromBytes(context.Background(), "test.txt", []byte("content"), "text/plain")
		require.NoError(t, err)

		exists, err := storage.Exists(context.Background(), "test.txt")
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("file does not exist", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "notexists-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		exists, err := storage.Exists(context.Background(), "nonexistent.txt")
		require.NoError(t, err)
		require.False(t, exists)
	})
}

func TestLocalStorage_GetMetadata(t *testing.T) {
	t.Run("successful metadata", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "metadata-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		content := []byte("metadata test content")
		err = storage.UploadFromBytes(context.Background(), "test.txt", content, "text/plain")
		require.NoError(t, err)

		metadata, err := storage.GetMetadata(context.Background(), "test.txt")
		require.NoError(t, err)
		require.NotNil(t, metadata)
		require.Equal(t, int64(len(content)), metadata.Size)
		require.NotEmpty(t, metadata.ETag)
		require.False(t, metadata.LastModified.IsZero())
	})

	t.Run("file not found", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "metadata-notfound-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		_, err = storage.GetMetadata(context.Background(), "nonexistent.txt")
		require.Error(t, err)
		require.Equal(t, ErrFileNotFound, err)
	})
}

func TestLocalStorage_GetPublicURL(t *testing.T) {
	t.Run("with base URL", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "url-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		url := storage.GetPublicURL("path/to/file.txt")
		require.Equal(t, "http://localhost:8080/files/path/to/file.txt", url)
	})

	t.Run("without base URL", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "nourl-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "")
		require.NoError(t, err)

		url := storage.GetPublicURL("file.txt")
		require.Empty(t, url)
	})
}

func TestLocalStorage_GetPresignedUploadURL(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "presigned-upload-test")
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalStorage(tmpDir, "http://localhost:8080")
	require.NoError(t, err)

	url, fields, err := storage.GetPresignedUploadURL(context.Background(), "upload/path.txt", 3600)
	require.NoError(t, err)
	require.NotEmpty(t, url)
	require.NotNil(t, fields)
	require.Equal(t, "upload/path.txt", fields["path"])
}

func TestLocalStorage_GetPresignedDownloadURL(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "presigned-download-test")
	defer os.RemoveAll(tmpDir)

	storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
	require.NoError(t, err)

	url, err := storage.GetPresignedDownloadURL(context.Background(), "download.txt", 3600)
	require.NoError(t, err)
	require.Equal(t, "http://localhost:8080/files/download.txt", url)
}

func TestLocalStorage_Copy(t *testing.T) {
	t.Run("successful copy", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "copy-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		content := []byte("copy test content")
		err = storage.UploadFromBytes(context.Background(), "source.txt", content, "text/plain")
		require.NoError(t, err)

		err = storage.Copy(context.Background(), "source.txt", "destination.txt")
		require.NoError(t, err)

		// Verify copy
		dstContent, err := storage.DownloadToBytes(context.Background(), "destination.txt")
		require.NoError(t, err)
		require.Equal(t, content, dstContent)

		// Verify original still exists
		srcContent, err := storage.DownloadToBytes(context.Background(), "source.txt")
		require.NoError(t, err)
		require.Equal(t, content, srcContent)
	})

	t.Run("copy non-existent file", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "copy-notfound-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		err = storage.Copy(context.Background(), "nonexistent.txt", "destination.txt")
		require.Error(t, err)
		require.Equal(t, ErrFileNotFound, err)
	})

	t.Run("copy creates directories", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "copy-dirs-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		err = storage.UploadFromBytes(context.Background(), "source.txt", []byte("content"), "text/plain")
		require.NoError(t, err)

		err = storage.Copy(context.Background(), "source.txt", "nested/path/destination.txt")
		require.NoError(t, err)

		// Verify copy exists
		_, err = storage.Exists(context.Background(), "nested/path/destination.txt")
		require.NoError(t, err)
		require.True(t, true)
	})
}

func TestLocalStorage_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent uploads", func(t *testing.T) {
		tmpDir := filepath.Join(os.TempDir(), "concurrent-test")
		defer os.RemoveAll(tmpDir)

		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080/files")
		require.NoError(t, err)

		// Upload multiple files concurrently
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(index int) {
				content := []byte("concurrent content")
				path := filepath.Join("concurrent", string(rune('A'+index)))
				err := storage.UploadFromBytes(context.Background(), path, content, "text/plain")
				require.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for alluploads to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

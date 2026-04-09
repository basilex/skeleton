package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalStorage implements StorageProvider using the local filesystem.
// Suitable for development and single-server deployments.
type LocalStorage struct {
	basePath string // Base directory for file storage
	baseURL  string // Base URL for public file access
}

// NewLocalStorage creates a new local storage provider.
func NewLocalStorage(basePath, baseURL string) (*LocalStorage, error) {
	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("create base path: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}, nil
}

// Upload stores data at the specified path.
func (s *LocalStorage) Upload(ctx context.Context, path string, reader io.Reader, contentType string) error {
	fullPath := filepath.Join(s.basePath, path)

	// Create directory if not exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	// Copy data
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("write data: %w", err)
	}

	return nil
}

// UploadFromBytes stores byte data at the specified path.
func (s *LocalStorage) UploadFromBytes(ctx context.Context, path string, data []byte, contentType string) error {
	fullPath := filepath.Join(s.basePath, path)

	// Create directory if not exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// Download retrieves data from the specified path.
func (s *LocalStorage) Download(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, path)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("open file: %w", err)
	}

	return file, nil
}

// DownloadToBytes retrieves all data into a byte slice.
func (s *LocalStorage) DownloadToBytes(ctx context.Context, path string) ([]byte, error) {
	fullPath := filepath.Join(s.basePath, path)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("read file: %w", err)
	}

	return data, nil
}

// Delete removes the file at the specified path.
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.basePath, path)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}

// Exists checks if a file exists at the specified path.
func (s *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(s.basePath, path)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("stat file: %w", err)
	}

	return true, nil
}

// GetMetadata returns metadata about the file.
func (s *LocalStorage) GetMetadata(ctx context.Context, path string) (*StorageMetadata, error) {
	fullPath := filepath.Join(s.basePath, path)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("stat file: %w", err)
	}

	// Calculate checksum
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	hash := sha256.Sum256(data)
	etag := hex.EncodeToString(hash[:])

	return &StorageMetadata{
		Size:         info.Size(),
		LastModified: info.ModTime(),
		ETag:         etag,
	}, nil
}

// GetPublicURL returns a public URL for the file.
func (s *LocalStorage) GetPublicURL(path string) string {
	if s.baseURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", s.baseURL, path)
}

// GetPresignedUploadURL generates a presigned URL for upload.
// For local storage, returns a simple upload endpoint URL.
func (s *LocalStorage) GetPresignedUploadURL(ctx context.Context, path string, expiresIn time.Duration) (string, map[string]string, error) {
	// Local storage doesn't support presigned URLs
	// Return a simple upload endpoint that the application can handle
	uploadURL := fmt.Sprintf("%s/upload/%s", s.baseURL, path)
	fields := map[string]string{
		"path": path,
	}

	return uploadURL, fields, nil
}

// GetPresignedDownloadURL generates a presigned URL for download.
func (s *LocalStorage) GetPresignedDownloadURL(ctx context.Context, path string, expiresIn time.Duration) (string, error) {
	// For local storage, return the public URL
	return s.GetPublicURL(path), nil
}

// Copy duplicates a file from src to dst.
func (s *LocalStorage) Copy(ctx context.Context, src, dst string) error {
	srcPath := filepath.Join(s.basePath, src)
	dstPath := filepath.Join(s.basePath, dst)

	// Create destination directory
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	// Open source file
	srcFile, err := os.Open(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return fmt.Errorf("open source: %w", err)
	}
	defer srcFile.Close()

	// Create destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer dstFile.Close()

	// Copy data
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy data: %w", err)
	}

	return nil
}

// Name returns the storage provider name.
func (s *LocalStorage) Name() string {
	return "local"
}

// ErrFileNotFound is returned when a file is not found.
var ErrFileNotFound = fmt.Errorf("file not found")

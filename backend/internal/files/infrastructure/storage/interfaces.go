package storage

import (
	"context"
	"io"
	"time"
)

// StorageMetadata contains metadata about a stored file.
type StorageMetadata struct {
	Size         int64     // File size in bytes
	ContentType  string    // MIME type
	LastModified time.Time // Last modification time
	ETag         string    // Entity tag (checksum)
}

// StorageProvider defines the interface for file storage backends.
// Implementations: LocalStorage, S3Storage, GCSStorage.
type StorageProvider interface {
	// Upload stores data at the specified path.
	Upload(ctx context.Context, path string, reader io.Reader, contentType string) error

	// UploadFromBytes stores byte data at the specified path.
	UploadFromBytes(ctx context.Context, path string, data []byte, contentType string) error

	// Download retrieves data from the specified path.
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// DownloadToBytes retrieves all data into a byte slice.
	DownloadToBytes(ctx context.Context, path string) ([]byte, error)

	// Delete removes the file at the specified path.
	Delete(ctx context.Context, path string) error

	// Exists checks if a file exists at the specified path.
	Exists(ctx context.Context, path string) (bool, error)

	// GetMetadata returns metadata about the file.
	GetMetadata(ctx context.Context, path string) (*StorageMetadata, error)

	// GetPublicURL returns a public URL for the file (if applicable).
	// Returns empty string for storage providers that don't support public URLs.
	GetPublicURL(path string) string

	// GetPresignedUploadURL generates a presigned URL for direct upload.
	// Returns upload URL, additional form fields, and expiration time.
	GetPresignedUploadURL(ctx context.Context, path string, expiresIn time.Duration) (url string, fields map[string]string, err error)

	// GetPresignedDownloadURL generates a presigned URL for download.
	GetPresignedDownloadURL(ctx context.Context, path string, expiresIn time.Duration) (string, error)

	// Copy duplicates a file from src to dst.
	Copy(ctx context.Context, src, dst string) error

	// Name returns the storage provider name (local, s3, gcs).
	Name() string
}

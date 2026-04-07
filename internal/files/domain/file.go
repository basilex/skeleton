package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
)

// StorageProvider represents the storage backend type.
type StorageProvider string

const (
	StorageS3    StorageProvider = "s3"    // AWS S3
	StorageGCS   StorageProvider = "gcs"   // Google Cloud Storage
	StorageLocal StorageProvider = "local" // Local filesystem
)

// AccessLevel represents file access permissions.
type AccessLevel string

const (
	AccessPublic     AccessLevel = "public"     // Anyone can access
	AccessPrivate    AccessLevel = "private"    // Only owner can access
	AccessRestricted AccessLevel = "restricted" // Specific permissions required
)

// FileMetadata contains additional file metadata.
type FileMetadata struct {
	Width      *int              // Image width in pixels
	Height     *int              // Image height in pixels
	Duration   *int              // Video/audio duration in seconds
	Pages      *int              // Document page count
	Thumbnail  *FileID           // Thumbnail file ID
	OriginalID *FileID           // Original file ID (for processed versions)
	Custom     map[string]string // Custom metadata
}

// File represents a file aggregate in the system.
// A file can be an uploaded document, image, video, or any other media.
// Files support metadata, expiration, access control, and processing operations.
type File struct {
	id              FileID
	ownerID         *domain.UserID // Optional owner (nullable for anonymous uploads)
	filename        string         // Original filename from upload
	storedName      string         // Generated unique storage name
	mimeType        string         // MIME type (e.g., "image/jpeg")
	size            int64          // Size in bytes
	path            string         // Storage path
	storageProvider StorageProvider
	checksum        string // SHA-256 hash
	metadata        FileMetadata
	accessLevel     AccessLevel
	uploadedAt      time.Time
	expiresAt       *time.Time
	processedAt     *time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

// NewFile creates a new File aggregate with validation.
// Parameters:
//   - ownerID: optional owner (nil for anonymous uploads)
//   - filename: original filename from upload
//   - mimeType: MIME type (e.g., "image/jpeg")
//   - size: file size in bytes
//   - provider: storage provider (local, s3, gcs)
//   - accessLevel: access permissions
func NewFile(
	ownerID *domain.UserID,
	filename string,
	mimeType string,
	size int64,
	provider StorageProvider,
	accessLevel AccessLevel,
) (*File, error) {
	if filename == "" {
		return nil, NewValidationError("filename", "filename cannot be empty")
	}
	if mimeType == "" {
		return nil, NewValidationError("mimeType", "MIME type cannot be empty")
	}
	if size < 0 {
		return nil, NewValidationError("size", "file size cannot be negative")
	}
	if !isValidStorageProvider(provider) {
		return nil, NewValidationError("storageProvider", "invalid storage provider")
	}
	if !isValidAccessLevel(accessLevel) {
		return nil, NewValidationError("accessLevel", "invalid access level")
	}

	now := time.Now()
	return &File{
		id:              NewFileID(),
		ownerID:         ownerID,
		filename:        filename,
		mimeType:        mimeType,
		size:            size,
		storageProvider: provider,
		accessLevel:     accessLevel,
		uploadedAt:      now,
		createdAt:       now,
		updatedAt:       now,
		metadata:        FileMetadata{Custom: make(map[string]string)},
	}, nil
}

// ReconstituteFile reconstructs a File from persistence.
// Used by repositories to create File from database records.
func ReconstituteFile(
	id FileID,
	ownerID *domain.UserID,
	filename string,
	storedName string,
	mimeType string,
	size int64,
	path string,
	provider StorageProvider,
	checksum string,
	metadata FileMetadata,
	accessLevel AccessLevel,
	uploadedAt time.Time,
	expiresAt *time.Time,
	processedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *File {
	return &File{
		id:              id,
		ownerID:         ownerID,
		filename:        filename,
		storedName:      storedName,
		mimeType:        mimeType,
		size:            size,
		path:            path,
		storageProvider: provider,
		checksum:        checksum,
		metadata:        metadata,
		accessLevel:     accessLevel,
		uploadedAt:      uploadedAt,
		expiresAt:       expiresAt,
		processedAt:     processedAt,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

// ID returns the file ID.
func (f *File) ID() FileID { return f.id }

// OwnerID returns the owner ID (nil for anonymous uploads).
func (f *File) OwnerID() *domain.UserID { return f.ownerID }

// Filename returns the original filename.
func (f *File) Filename() string { return f.filename }

// StoredName returns the generated storage name.
func (f *File) StoredName() string { return f.storedName }

// MimeType returns the MIME type.
func (f *File) MimeType() string { return f.mimeType }

// Size returns the file size in bytes.
func (f *File) Size() int64 { return f.size }

// Path returns the storage path.
func (f *File) Path() string { return f.path }

// StorageProvider returns the storage provider.
func (f *File) StorageProvider() StorageProvider { return f.storageProvider }

// Checksum returns the SHA-256 checksum.
func (f *File) Checksum() string { return f.checksum }

// Metadata returns the file metadata.
func (f *File) Metadata() FileMetadata { return f.metadata }

// AccessLevel returns the access level.
func (f *File) AccessLevel() AccessLevel { return f.accessLevel }

// UploadedAt returns the upload timestamp.
func (f *File) UploadedAt() time.Time { return f.uploadedAt }

// ExpiresAt returns the expiration timestamp (nil if no expiration).
func (f *File) ExpiresAt() *time.Time { return f.expiresAt }

// ProcessedAt returns the processing completion timestamp (nil if not processed).
func (f *File) ProcessedAt() *time.Time { return f.processedAt }

// CreatedAt returns the creation timestamp.
func (f *File) CreatedAt() time.Time { return f.createdAt }

// UpdatedAt returns the last update timestamp.
func (f *File) UpdatedAt() time.Time { return f.updatedAt }

// SetPath sets the storage path and generates the stored name.
// Called after successful upload to storage.
func (f *File) SetPath(path string) error {
	if path == "" {
		return NewValidationError("path", "path cannot be empty")
	}
	f.path = path
	f.storedName = filepath.Base(path)
	f.updatedAt = time.Now()
	return nil
}

// SetChecksum calculates and sets the SHA-256 checksum from data.
func (f *File) SetChecksum(data []byte) {
	hash := sha256.Sum256(data)
	f.checksum = hex.EncodeToString(hash[:])
	f.updatedAt = time.Now()
}

// SetChecksumFromHex sets the checksum from a hex string.
func (f *File) SetChecksumFromHex(hexChecksum string) error {
	if hexChecksum == "" {
		return NewValidationError("checksum", "checksum cannot be empty")
	}
	// Validate hex format
	if len(hexChecksum) != 64 {
		return NewValidationError("checksum", "invalid checksum format")
	}
	f.checksum = hexChecksum
	f.updatedAt = time.Now()
	return nil
}

// SetMetadata sets the file metadata.
func (f *File) SetMetadata(metadata FileMetadata) error {
	f.metadata = metadata
	if f.metadata.Custom == nil {
		f.metadata.Custom = make(map[string]string)
	}
	f.updatedAt = time.Now()
	return nil
}

// SetExpiration sets the file expiration time.
func (f *File) SetExpiration(expiresAt time.Time) error {
	if expiresAt.Before(time.Now()) {
		return NewValidationError("expiresAt", "expiration time cannot be in the past")
	}
	f.expiresAt = &expiresAt
	f.updatedAt = time.Now()
	return nil
}

// SetProcessed marks the file as processed.
func (f *File) SetProcessed() {
	now := time.Now()
	f.processedAt = &now
	f.updatedAt = now
}

// IsExpired returns true if the file has expired.
func (f *File) IsExpired() bool {
	if f.expiresAt == nil {
		return false
	}
	return time.Now().After(*f.expiresAt)
}

// IsImage returns true if the file is an image.
func (f *File) IsImage() bool {
	return strings.HasPrefix(f.mimeType, "image/")
}

// IsVideo returns true if the file is a video.
func (f *File) IsVideo() bool {
	return strings.HasPrefix(f.mimeType, "video/")
}

// IsAudio returns true if the file is audio.
func (f *File) IsAudio() bool {
	return strings.HasPrefix(f.mimeType, "audio/")
}

// IsDocument returns true if the file is a document.
func (f *File) IsDocument() bool {
	docTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument",
		"application/vnd.ms-",
		"text/",
	}
	for _, t := range docTypes {
		if strings.HasPrefix(f.mimeType, t) {
			return true
		}
	}
	return false
}

// CanAccess returns true if the given user can access the file.
func (f *File) CanAccess(userID *domain.UserID) bool {
	// Expired files cannot be accessed
	if f.IsExpired() {
		return false
	}

	// Public files can be accessed by anyone
	if f.accessLevel == AccessPublic {
		return true
	}

	// Anonymous users can only access public files
	if userID == nil {
		return false
	}

	// Private files can only be accessed by owner
	if f.accessLevel == AccessPrivate {
		return f.ownerID != nil && *f.ownerID == *userID
	}

	// Restricted files require specific permissions (checked at application layer)
	// For now, only owner can access
	if f.accessLevel == AccessRestricted {
		return f.ownerID != nil && *f.ownerID == *userID
	}

	return false
}

// GeneratePath generates a storage path based on file ID and storage provider.
func (f *File) GeneratePath() string {
	ext := filepath.Ext(f.filename)
	return fmt.Sprintf("%s/%s/%s",
		f.storageProvider,
		f.id.String()[:2], // First 2 chars for directory sharding
		f.id.String()+ext,
	)
}

// GenerateStoredName generates a unique storage name.
func (f *File) GenerateStoredName() string {
	ext := filepath.Ext(f.filename)
	return f.id.String() + ext
}

// isValidStorageProvider validates the storage provider.
func isValidStorageProvider(provider StorageProvider) bool {
	switch provider {
	case StorageS3, StorageGCS, StorageLocal:
		return true
	default:
		return false
	}
}

// isValidAccessLevel validates the access level.
func isValidAccessLevel(level AccessLevel) bool {
	switch level {
	case AccessPublic, AccessPrivate, AccessRestricted:
		return true
	default:
		return false
	}
}

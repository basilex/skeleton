package domain

import (
	"testing"
	"time"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/stretchr/testify/require"
)

func TestNewFileUpload(t *testing.T) {
	userID := identityDomain.UserID("user-123")
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)
	ttl := 1 * time.Hour

	t.Run("valid upload", func(t *testing.T) {
		upload, err := NewFileUpload(file, ttl)
		require.NoError(t, err)
		require.NotNil(t, upload)
		require.NotEmpty(t, upload.ID())
		require.Equal(t, file, upload.File())
		require.Equal(t, UploadPending, upload.Status())
		require.False(t, upload.CreatedAt().IsZero())
		require.True(t, upload.ExpiresAt().After(time.Now()))
	})

	t.Run("nil file", func(t *testing.T) {
		_, err := NewFileUpload(nil, ttl)
		require.Error(t, err)
		require.Contains(t, err.Error(), "file cannot be nil")
	})

	t.Run("zero TTL", func(t *testing.T) {
		_, err := NewFileUpload(file, 0)
		require.Error(t, err)
		require.Contains(t, err.Error(), "TTL must be positive")
	})

	t.Run("negative TTL", func(t *testing.T) {
		_, err := NewFileUpload(file, -1*time.Hour)
		require.Error(t, err)
		require.Contains(t, err.Error(), "TTL must be positive")
	})
}

func TestFileUploadSetUploadURL(t *testing.T) {
	userID := identityDomain.UserID("user-123")
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)
	upload, _ := NewFileUpload(file, 1*time.Hour)

	url := "https://storage.example.com/upload/123"
	fields := map[string]string{
		"key":       "uploads/test.jpg",
		"policy":    "base64policy",
		"signature": "signature",
	}

	upload.SetUploadURL(url, fields)
	require.Equal(t, url, upload.UploadURL())
	require.Equal(t, fields, upload.Fields())
}

func TestFileUploadSetUploadURLWithNilFields(t *testing.T) {
	userID := identityDomain.UserID("user-123")
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)
	upload, _ := NewFileUpload(file, 1*time.Hour)

	url := "https://storage.example.com/upload/123"
	upload.SetUploadURL(url, nil)
	require.Equal(t, url, upload.UploadURL())
}

func TestFileUploadMarkCompleted(t *testing.T) {
	userID := identityDomain.UserID("user-123")
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)

	t.Run("from pending", func(t *testing.T) {
		upload, _ := NewFileUpload(file, 1*time.Hour)
		err := upload.MarkCompleted()
		require.NoError(t, err)
		require.Equal(t, UploadCompleted, upload.Status())
	})

	t.Run("already completed", func(t *testing.T) {
		upload, _ := NewFileUpload(file, 1*time.Hour)
		_ = upload.MarkCompleted()
		err := upload.MarkCompleted()
		require.Error(t, err)
		require.Equal(t, ErrUploadAlreadyCompleted, err)
	})

	t.Run("from expired", func(t *testing.T) {
		upload, _ := NewFileUpload(file, 1*time.Hour)
		upload.MarkFailed()
		err := upload.MarkCompleted()
		require.Error(t, err)
		require.Equal(t, ErrUploadNotCompleted, err)
	})
}

func TestFileUploadMarkFailed(t *testing.T) {
	userID := identityDomain.UserID("user-123")
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)
	upload, _ := NewFileUpload(file, 1*time.Hour)

	upload.MarkFailed()
	require.Equal(t, UploadFailed, upload.Status())
}

func TestFileUploadIsExpired(t *testing.T) {
	userID := identityDomain.UserID("user-123")
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)

	t.Run("not expired", func(t *testing.T) {
		upload, _ := NewFileUpload(file, 1*time.Hour)
		require.False(t, upload.IsExpired())
	})

	t.Run("expired", func(t *testing.T) {
		upload := ReconstituteFileUpload(
			NewUploadID(),
			file,
			"",
			nil,
			UploadPending,
			time.Now().Add(-1*time.Hour),
			time.Now(),
		)
		require.True(t, upload.IsExpired())
	})
}

func TestFileUploadCanUpload(t *testing.T) {
	userID := identityDomain.UserID("user-123")
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)

	t.Run("pending and not expired", func(t *testing.T) {
		upload, _ := NewFileUpload(file, 1*time.Hour)
		require.True(t, upload.CanUpload())
	})

	t.Run("pending but expired", func(t *testing.T) {
		upload := ReconstituteFileUpload(
			NewUploadID(),
			file,
			"",
			nil,
			UploadPending,
			time.Now().Add(-1*time.Hour),
			time.Now(),
		)
		require.False(t, upload.CanUpload())
	})

	t.Run("completed", func(t *testing.T) {
		upload, _ := NewFileUpload(file, 1*time.Hour)
		_ = upload.MarkCompleted()
		require.False(t, upload.CanUpload())
	})

	t.Run("failed", func(t *testing.T) {
		upload, _ := NewFileUpload(file, 1*time.Hour)
		upload.MarkFailed()
		require.False(t, upload.CanUpload())
	})
}

func TestFileUploadFields(t *testing.T) {
	userID := identityDomain.UserID("user-123")
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)
	upload, _ := NewFileUpload(file, 1*time.Hour)

	fields := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	upload.SetUploadURL("https://example.com/upload", fields)

	retrievedFields := upload.Fields()
	require.Equal(t, fields, retrievedFields)

	retrievedFields["key3"] = "value3"
	require.NotContains(t, upload.Fields(), "key3")
}

func TestReconstituteFileUpload(t *testing.T) {
	userID := identityDomain.UserID("user-123")
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)

	fields := map[string]string{
		"key":  "value",
		"key2": "value2",
	}

	upload := ReconstituteFileUpload(
		UploadID("upload-123"),
		file,
		"https://storage.example.com/upload",
		fields,
		UploadCompleted,
		expiresAt,
		now,
	)

	require.Equal(t, UploadID("upload-123"), upload.ID())
	require.Equal(t, file, upload.File())
	require.Equal(t, "https://storage.example.com/upload", upload.UploadURL())
	require.Equal(t, fields, upload.Fields())
	require.Equal(t, UploadCompleted, upload.Status())
	require.Equal(t, expiresAt.Unix(), upload.ExpiresAt().Unix())
	require.Equal(t, now.Unix(), upload.CreatedAt().Unix())
}

func TestReconstituteFileUploadWithNilFields(t *testing.T) {
	userID := identityDomain.UserID("user-123")
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)
	now := time.Now()

	upload := ReconstituteFileUpload(
		UploadID("upload-123"),
		file,
		"https://storage.example.com/upload",
		nil,
		UploadPending,
		now.Add(1*time.Hour),
		now,
	)

	require.NotNil(t, upload.Fields())
	require.Empty(t, upload.Fields())
}

func TestUploadStatusString(t *testing.T) {
	require.Equal(t, "pending", string(UploadPending))
	require.Equal(t, "completed", string(UploadCompleted))
	require.Equal(t, "failed", string(UploadFailed))
	require.Equal(t, "expired", string(UploadExpired))
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		name   string
		status UploadStatus
		valid  bool
	}{
		{name: "pending", status: UploadPending, valid: true},
		{name: "completed", status: UploadCompleted, valid: true},
		{name: "failed", status: UploadFailed, valid: true},
		{name: "expired", status: UploadExpired, valid: true},
		{name: "invalid", status: UploadStatus("invalid"), valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.valid, IsValidStatus(tt.status))
		})
	}
}

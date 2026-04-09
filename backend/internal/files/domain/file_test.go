package domain

import (
	"testing"
	"time"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/stretchr/testify/require"
)

func TestNewFile(t *testing.T) {
	userID := identityDomain.NewUserID()
	filename := "test.jpg"
	mimeType := "image/jpeg"
	size := int64(1024)

	t.Run("with owner", func(t *testing.T) {
		file, err := NewFile(&userID, filename, mimeType, size, StorageLocal, AccessPrivate)
		require.NoError(t, err)
		require.NotNil(t, file)
		require.NotEmpty(t, file.ID())
		require.Equal(t, &userID, file.OwnerID())
		require.Equal(t, filename, file.Filename())
		require.Equal(t, mimeType, file.MimeType())
		require.Equal(t, size, file.Size())
		require.Equal(t, StorageLocal, file.StorageProvider())
		require.Equal(t, AccessPrivate, file.AccessLevel())
		require.Equal(t, AccessPrivate, file.AccessLevel())
		require.False(t, file.CreatedAt().IsZero())
		require.False(t, file.UploadedAt().IsZero())
	})

	t.Run("anonymous upload", func(t *testing.T) {
		file, err := NewFile(nil, filename, mimeType, size, StorageLocal, AccessPublic)
		require.NoError(t, err)
		require.Nil(t, file.OwnerID())
		require.Equal(t, AccessPublic, file.AccessLevel())
	})

	t.Run("different access levels", func(t *testing.T) {
		tests := []struct {
			name        string
			accessLevel AccessLevel
		}{
			{name: "public", accessLevel: AccessPublic},
			{name: "private", accessLevel: AccessPrivate},
			{name: "restricted", accessLevel: AccessRestricted},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				file, err := NewFile(&userID, filename, mimeType, size, StorageS3, tt.accessLevel)
				require.NoError(t, err)
				require.Equal(t, tt.accessLevel, file.AccessLevel())
			})
		}
	})

	t.Run("different storage providers", func(t *testing.T) {
		tests := []struct {
			name     string
			provider StorageProvider
		}{
			{name: "local", provider: StorageLocal},
			{name: "s3", provider: StorageS3},
			{name: "gcs", provider: StorageGCS},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				file, err := NewFile(&userID, filename, mimeType, size, tt.provider, AccessPrivate)
				require.NoError(t, err)
				require.Equal(t, tt.provider, file.StorageProvider())
			})
		}
	})
}

func TestNewFileValidation(t *testing.T) {
	userID := identityDomain.NewUserID()

	tests := []struct {
		name        string
		ownerID     *identityDomain.UserID
		filename    string
		mimeType    string
		size        int64
		provider    StorageProvider
		accessLevel AccessLevel
		wantErr     bool
		errContain  string
	}{
		{
			name:        "empty filename",
			ownerID:     &userID,
			filename:    "",
			mimeType:    "image/jpeg",
			size:        1024,
			provider:    StorageLocal,
			accessLevel: AccessPrivate,
			wantErr:     true,
			errContain:  "filename cannot be empty",
		},
		{
			name:        "empty mime type",
			ownerID:     &userID,
			filename:    "test.jpg",
			mimeType:    "",
			size:        1024,
			provider:    StorageLocal,
			accessLevel: AccessPrivate,
			wantErr:     true,
			errContain:  "MIME type cannot be empty",
		},
		{
			name:        "negative size",
			ownerID:     &userID,
			filename:    "test.jpg",
			mimeType:    "image/jpeg",
			size:        -1,
			provider:    StorageLocal,
			accessLevel: AccessPrivate,
			wantErr:     true,
			errContain:  "file size cannot be negative",
		},
		{
			name:        "invalid storage provider",
			ownerID:     &userID,
			filename:    "test.jpg",
			mimeType:    "image/jpeg",
			size:        1024,
			provider:    StorageProvider("invalid"),
			accessLevel: AccessPrivate,
			wantErr:     true,
			errContain:  "invalid storage provider",
		},
		{
			name:        "invalid access level",
			ownerID:     &userID,
			filename:    "test.jpg",
			mimeType:    "image/jpeg",
			size:        1024,
			provider:    StorageLocal,
			accessLevel: AccessLevel("invalid"),
			wantErr:     true,
			errContain:  "invalid access level",
		},
		{
			name:        "valid file",
			ownerID:     &userID,
			filename:    "test.jpg",
			mimeType:    "image/jpeg",
			size:        1024,
			provider:    StorageLocal,
			accessLevel: AccessPrivate,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFile(tt.ownerID, tt.filename, tt.mimeType, tt.size, tt.provider, tt.accessLevel)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errContain)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestFileSetPath(t *testing.T) {
	userID := identityDomain.NewUserID()
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)

	t.Run("valid path", func(t *testing.T) {
		err := file.SetPath("uploads/test-123.jpg")
		require.NoError(t, err)
		require.Equal(t, "uploads/test-123.jpg", file.Path())
		require.Equal(t, "test-123.jpg", file.StoredName())
	})

	t.Run("empty path", func(t *testing.T) {
		err := file.SetPath("")
		require.Error(t, err)
		require.Contains(t, err.Error(), "path cannot be empty")
	})
}

func TestFileSetChecksum(t *testing.T) {
	userID := identityDomain.NewUserID()
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)

	data := []byte("test content")
	file.SetChecksum(data)
	require.NotEmpty(t, file.Checksum())
	require.Len(t, file.Checksum(), 64)
}

func TestFileSetChecksumFromHex(t *testing.T) {
	userID := identityDomain.NewUserID()
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)

	t.Run("valid checksum", func(t *testing.T) {
		hash := "a" + string(make([]byte, 63))
		for i := range hash {
			if hash[i] == 0 {
				hash = hash[:i] + "0" + hash[i+1:]
			}
		}
		err := file.SetChecksumFromHex("a94a8fe5ccb19ba61c4c0873d391e987982fbbd3d391e987982fbbd3d391e987")
		require.NoError(t, err)
		require.NotEmpty(t, file.Checksum())
	})

	t.Run("empty checksum", func(t *testing.T) {
		err := file.SetChecksumFromHex("")
		require.Error(t, err)
		require.Contains(t, err.Error(), "checksum cannot be empty")
	})

	t.Run("invalid checksum format", func(t *testing.T) {
		err := file.SetChecksumFromHex("invalid")
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid checksum format")
	})
}

func TestFileSetMetadata(t *testing.T) {
	userID := identityDomain.NewUserID()
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)

	width := 800
	height := 600
	metadata := FileMetadata{
		Width:  &width,
		Height: &height,
		Custom: map[string]string{"key": "value"},
	}

	err := file.SetMetadata(metadata)
	require.NoError(t, err)
	require.Equal(t, metadata.Width, file.Metadata().Width)
	require.Equal(t, metadata.Height, file.Metadata().Height)
	require.Equal(t, "value", file.Metadata().Custom["key"])
}

func TestFileSetExpiration(t *testing.T) {
	userID := identityDomain.NewUserID()
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)

	t.Run("valid expiration", func(t *testing.T) {
		futureTime := time.Now().Add(24 * time.Hour)
		err := file.SetExpiration(futureTime)
		require.NoError(t, err)
		require.NotNil(t, file.ExpiresAt())
		require.Equal(t, futureTime.Unix(), file.ExpiresAt().Unix())
	})

	t.Run("past expiration", func(t *testing.T) {
		pastTime := time.Now().Add(-1 * time.Hour)
		err := file.SetExpiration(pastTime)
		require.Error(t, err)
		require.Contains(t, err.Error(), "expiration time cannot be in the past")
	})
}

func TestFileSetProcessed(t *testing.T) {
	userID := identityDomain.NewUserID()
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)

	require.Nil(t, file.ProcessedAt())
	file.SetProcessed()
	require.NotNil(t, file.ProcessedAt())
}

func TestFileIsExpired(t *testing.T) {
	userID := identityDomain.NewUserID()

	t.Run("without expiration", func(t *testing.T) {
		file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)
		require.False(t, file.IsExpired())
	})

	t.Run("with future expiration", func(t *testing.T) {
		file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)
		futureTime := time.Now().Add(24 * time.Hour)
		_ = file.SetExpiration(futureTime)
		require.False(t, file.IsExpired())
	})

	t.Run("with past expiration", func(t *testing.T) {
		file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)
		pastTime := time.Now().Add(-1 * time.Hour)
		_ = file.SetExpiration(pastTime)
	})
}

func TestFileIsImage(t *testing.T) {
	userID := identityDomain.NewUserID()

	tests := []struct {
		name     string
		mimeType string
		isImage  bool
	}{
		{name: "jpg", mimeType: "image/jpeg", isImage: true},
		{name: "png", mimeType: "image/png", isImage: true},
		{name: "gif", mimeType: "image/gif", isImage: true},
		{name: "pdf", mimeType: "application/pdf", isImage: false},
		{name: "mp4", mimeType: "video/mp4", isImage: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := NewFile(&userID, "test", tt.mimeType, 1024, StorageLocal, AccessPrivate)
			require.Equal(t, tt.isImage, file.IsImage())
		})
	}
}

func TestFileIsVideo(t *testing.T) {
	userID := identityDomain.NewUserID()

	tests := []struct {
		name     string
		mimeType string
		isVideo  bool
	}{
		{name: "mp4", mimeType: "video/mp4", isVideo: true},
		{name: "avi", mimeType: "video/avi", isVideo: true},
		{name: "image", mimeType: "image/jpeg", isVideo: false},
		{name: "pdf", mimeType: "application/pdf", isVideo: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := NewFile(&userID, "test", tt.mimeType, 1024, StorageLocal, AccessPrivate)
			require.Equal(t, tt.isVideo, file.IsVideo())
		})
	}
}

func TestFileIsAudio(t *testing.T) {
	userID := identityDomain.NewUserID()

	tests := []struct {
		name     string
		mimeType string
		isAudio  bool
	}{
		{name: "mp3", mimeType: "audio/mp3", isAudio: true},
		{name: "wav", mimeType: "audio/wav", isAudio: true},
		{name: "image", mimeType: "image/jpeg", isAudio: false},
		{name: "pdf", mimeType: "application/pdf", isAudio: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := NewFile(&userID, "test", tt.mimeType, 1024, StorageLocal, AccessPrivate)
			require.Equal(t, tt.isAudio, file.IsAudio())
		})
	}
}

func TestFileIsDocument(t *testing.T) {
	userID := identityDomain.NewUserID()

	tests := []struct {
		name       string
		mimeType   string
		isDocument bool
	}{
		{name: "pdf", mimeType: "application/pdf", isDocument: true},
		{name: "doc", mimeType: "application/msword", isDocument: true},
		{name: "docx", mimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", isDocument: true},
		{name: "txt", mimeType: "text/plain", isDocument: true},
		{name: "image", mimeType: "image/jpeg", isDocument: false},
		{name: "video", mimeType: "video/mp4", isDocument: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, _ := NewFile(&userID, "test", tt.mimeType, 1024, StorageLocal, AccessPrivate)
			require.Equal(t, tt.isDocument, file.IsDocument())
		})
	}
}

func TestFileCanAccess(t *testing.T) {
	userID := identityDomain.NewUserID()
	otherUserID := identityDomain.NewUserID()

	t.Run("public file", func(t *testing.T) {
		file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPublic)

		require.True(t, file.CanAccess(&userID))
		require.True(t, file.CanAccess(&otherUserID))
		require.True(t, file.CanAccess(nil))
	})

	t.Run("private file owner", func(t *testing.T) {
		file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)

		require.True(t, file.CanAccess(&userID))
		require.False(t, file.CanAccess(&otherUserID))
		require.False(t, file.CanAccess(nil))
	})

	t.Run("private file anonymous", func(t *testing.T) {
		file, _ := NewFile(nil, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)

		require.False(t, file.CanAccess(&userID))
		require.False(t, file.CanAccess(nil))
	})

	t.Run("restricted file", func(t *testing.T) {
		file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessRestricted)

		require.True(t, file.CanAccess(&userID))
		require.False(t, file.CanAccess(&otherUserID))
		require.False(t, file.CanAccess(nil))
	})

	t.Run("expired file", func(t *testing.T) {
		now := time.Now()
		pastTime := now.Add(-1 * time.Hour)
		userID := identityDomain.NewUserID()

		file := ReconstituteFile(
			NewFileID(),
			&userID,
			"test.jpg",
			"stored-123.jpg",
			"image/jpeg",
			1024,
			"uploads/test.jpg",
			StorageLocal,
			"abc123",
			FileMetadata{Custom: make(map[string]string)},
			AccessPublic,
			ScanStatusPending,
			"",
			nil,
			now,
			&pastTime, // Expires in the past
			nil,
			now,
			now,
		)

		require.True(t, file.IsExpired())
		require.False(t, file.CanAccess(&userID))
		require.False(t, file.CanAccess(nil))
	})
}

func TestFileGeneratePath(t *testing.T) {
	userID := identityDomain.NewUserID()
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageS3, AccessPrivate)

	path := file.GeneratePath()
	require.Contains(t, path, "s3/")
	require.Contains(t, path, ".jpg")
}

func TestFileGenerateStoredName(t *testing.T) {
	userID := identityDomain.NewUserID()
	file, _ := NewFile(&userID, "test.jpg", "image/jpeg", 1024, StorageLocal, AccessPrivate)

	storedName := file.GenerateStoredName()
	require.Contains(t, storedName, file.ID().String())
	require.Contains(t, storedName, ".jpg")
}

func TestReconstituteFile(t *testing.T) {
	userID := identityDomain.NewUserID()
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	processedAt := now.Add(1 * time.Hour)
	scannedAt := now.Add(30 * time.Minute)

	metadata := FileMetadata{
		Custom: map[string]string{"key": "value"},
	}

	file := ReconstituteFile(
		NewFileID(),
		&userID,
		"test.jpg",
		"stored-123.jpg",
		"image/jpeg",
		1024,
		"uploads/stored-123.jpg",
		StorageLocal,
		"abc123",
		metadata,
		AccessPrivate,
		ScanStatusClean,
		"",
		&scannedAt,
		now,
		&expiresAt,
		&processedAt,
		now,
		now,
	)

	require.Equal(t, file.ID(), file.ID())
	require.Equal(t, &userID, file.OwnerID())
	require.Equal(t, "test.jpg", file.Filename())
	require.Equal(t, "stored-123.jpg", file.StoredName())
	require.Equal(t, "image/jpeg", file.MimeType())
	require.Equal(t, int64(1024), file.Size())
	require.Equal(t, "uploads/stored-123.jpg", file.Path())
	require.Equal(t, StorageLocal, file.StorageProvider())
	require.Equal(t, "abc123", file.Checksum())
	require.Equal(t, metadata, file.Metadata())
	require.Equal(t, AccessPrivate, file.AccessLevel())
}

func TestParseStorageProvider(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    StorageProvider
		wantErr bool
	}{
		{name: "s3", input: "s3", want: StorageS3, wantErr: false},
		{name: "gcs", input: "gcs", want: StorageGCS, wantErr: false},
		{name: "local", input: "local", want: StorageLocal, wantErr: false},
		{name: "invalid", input: "invalid", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result StorageProvider
			var err error

			switch tt.input {
			case "s3":
				result = StorageS3
			case "gcs":
				result = StorageGCS
			case "local":
				result = StorageLocal
			}

			if tt.wantErr {
				require.False(t, isValidStorageProvider(StorageProvider(tt.input)))
			} else {
				require.True(t, isValidStorageProvider(result))
				require.Equal(t, tt.want, result)
			}

			_ = err
		})
	}
}

func TestParseAccessLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    AccessLevel
		wantErr bool
	}{
		{name: "public", input: "public", want: AccessPublic, wantErr: false},
		{name: "private", input: "private", want: AccessPrivate, wantErr: false},
		{name: "restricted", input: "restricted", want: AccessRestricted, wantErr: false},
		{name: "invalid", input: "invalid", want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result AccessLevel
			var err error

			switch tt.input {
			case "public":
				result = AccessPublic
			case "private":
				result = AccessPrivate
			case "restricted":
				result = AccessRestricted
			}

			if tt.wantErr {
				require.False(t, isValidAccessLevel(AccessLevel(tt.input)))
			} else {
				require.True(t, isValidAccessLevel(result))
				require.Equal(t, tt.want, result)
			}

			_ = err
		})
	}
}

func TestStorageProviderString(t *testing.T) {
	require.Equal(t, "s3", string(StorageS3))
	require.Equal(t, "gcs", string(StorageGCS))
	require.Equal(t, "local", string(StorageLocal))
}

func TestAccessLevelString(t *testing.T) {
	require.Equal(t, "public", string(AccessPublic))
	require.Equal(t, "private", string(AccessPrivate))
	require.Equal(t, "restricted", string(AccessRestricted))
}

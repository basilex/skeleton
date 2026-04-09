package domain

import (
	"testing"
	"time"

	identityDomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/stretchr/testify/require"
)

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}

func TestFile_StartScan(t *testing.T) {
	userID := identityDomain.NewUserID()
	file, _ := NewFile(&userID, "test.pdf", "application/pdf", 1024, StorageLocal, AccessPrivate)

	require.Equal(t, ScanStatusPending, file.ScanStatus())

	t.Run("start scan", func(t *testing.T) {
		err := file.StartScan()
		require.NoError(t, err)
		require.Equal(t, ScanStatusScanning, file.ScanStatus())
	})

	t.Run("cannot start scan when already scanning", func(t *testing.T) {
		err := file.StartScan()
		require.Error(t, err)
		require.Contains(t, err.Error(), "already being scanned")
	})
}

func TestFile_MarkClean(t *testing.T) {
	userID := identityDomain.NewUserID()

	t.Run("mark clean after scan", func(t *testing.T) {
		file, _ := NewFile(&userID, "test.pdf", "application/pdf", 1024, StorageLocal, AccessPrivate)
		_ = file.StartScan()

		err := file.MarkClean()
		require.NoError(t, err)
		require.Equal(t, ScanStatusClean, file.ScanStatus())
		require.NotNil(t, file.ScannedAt())
		require.True(t, file.IsScanned())
		require.True(t, file.IsClean())
		require.False(t, file.IsInfected())
	})

	t.Run("cannot mark clean without scanning", func(t *testing.T) {
		file, _ := NewFile(&userID, "test.pdf", "application/pdf", 1024, StorageLocal, AccessPrivate)

		err := file.MarkClean()
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be in scanning status")
	})
}

func TestFile_MarkInfected(t *testing.T) {
	userID := identityDomain.NewUserID()

	t.Run("mark infected after scan", func(t *testing.T) {
		file, _ := NewFile(&userID, "test.pdf", "application/pdf", 1024, StorageLocal, AccessPrivate)
		_ = file.StartScan()

		err := file.MarkInfected("Trojan.GenericKD.12345")
		require.NoError(t, err)
		require.Equal(t, ScanStatusInfected, file.ScanStatus())
		require.NotNil(t, file.ScannedAt())
		require.Equal(t, "Trojan.GenericKD.12345", file.ThreatInfo())
		require.True(t, file.IsScanned())
		require.False(t, file.IsClean())
		require.True(t, file.IsInfected())
	})

	t.Run("cannot mark infected without scanning", func(t *testing.T) {
		file, _ := NewFile(&userID, "test.pdf", "application/pdf", 1024, StorageLocal, AccessPrivate)

		err := file.MarkInfected("test")
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be in scanning status")
	})
}

func TestFile_CanBeDownloaded(t *testing.T) {
	userID := identityDomain.NewUserID()

	t.Run("cannot download pending file", func(t *testing.T) {
		file, _ := NewFile(&userID, "test.pdf", "application/pdf", 1024, StorageLocal, AccessPrivate)
		require.False(t, file.CanBeDownloaded())
	})

	t.Run("cannot download scanning file", func(t *testing.T) {
		file, _ := NewFile(&userID, "test.pdf", "application/pdf", 1024, StorageLocal, AccessPrivate)
		_ = file.StartScan()
		require.False(t, file.CanBeDownloaded())
	})

	t.Run("cannot download infected file", func(t *testing.T) {
		file, _ := NewFile(&userID, "test.pdf", "application/pdf", 1024, StorageLocal, AccessPrivate)
		_ = file.StartScan()
		_ = file.MarkInfected("trojan")
		require.False(t, file.CanBeDownloaded())
	})

	t.Run("can download clean file", func(t *testing.T) {
		file, _ := NewFile(&userID, "test.pdf", "application/pdf", 1024, StorageLocal, AccessPrivate)
		_ = file.StartScan()
		_ = file.MarkClean()
		require.True(t, file.CanBeDownloaded())
	})
}

func TestFileTypePolicy(t *testing.T) {
	policy := DefaultFileTypePolicy()

	t.Run("validate allowed image type", func(t *testing.T) {
		err := policy.ValidateMimeType("image/jpeg")
		require.NoError(t, err)
	})

	t.Run("validate allowed document type", func(t *testing.T) {
		err := policy.ValidateMimeType("application/pdf")
		require.NoError(t, err)
	})

	t.Run("reject disallowed type", func(t *testing.T) {
		err := policy.ValidateMimeType("application/x-executable")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidFileType)
	})

	t.Run("validate size within limit", func(t *testing.T) {
		err := policy.ValidateSize(50 * 1024 * 1024) // 50MB
		require.NoError(t, err)
	})

	t.Run("reject size over limit", func(t *testing.T) {
		err := policy.ValidateSize(150 * 1024 * 1024) // 150MB (over 100MB limit)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrFileTooLarge)
	})

	t.Run("is type allowed", func(t *testing.T) {
		require.True(t, policy.IsTypeAllowed("image/png"))
		require.False(t, policy.IsTypeAllowed("application/x-executable"))
	})
}

func TestNewFileTypePolicy(t *testing.T) {
	t.Run("custom allowed types", func(t *testing.T) {
		policy := NewFileTypePolicy([]string{"image/jpeg", "image/png"}, 10*1024*1024)
		require.NoError(t, policy.ValidateMimeType("image/jpeg"))
		require.NoError(t, policy.ValidateMimeType("image/png"))
		require.Error(t, policy.ValidateMimeType("image/gif"))
	})

	t.Run("validate with policy", func(t *testing.T) {
		policy := NewFileTypePolicy([]string{"application/pdf"}, 10*1024*1024)

		err := policy.Validate("application/pdf", 5*1024*1024)
		require.NoError(t, err)

		err = policy.Validate("image/jpeg", 5*1024*1024)
		require.Error(t, err)

		err = policy.Validate("application/pdf", 20*1024*1024)
		require.Error(t, err)
	})
}

func TestReconstituteFile_WithScanStatus(t *testing.T) {
	userID := identityDomain.NewUserID()
	id := NewFileID()
	scannedAt := parseTime("2024-01-01T12:00:00Z")

	file := ReconstituteFile(
		id,
		&userID,
		"test.pdf",
		"stored-name.pdf",
		"application/pdf",
		1024,
		"/path/to/file",
		StorageLocal,
		"checksum123",
		FileMetadata{},
		AccessPrivate,
		ScanStatusClean,
		"",
		&scannedAt,
		scannedAt,
		nil,
		nil,
		scannedAt,
		scannedAt,
	)

	require.Equal(t, ScanStatusClean, file.ScanStatus())
	require.NotNil(t, file.ScannedAt())
	require.True(t, file.IsClean())
	require.True(t, file.CanBeDownloaded())
}

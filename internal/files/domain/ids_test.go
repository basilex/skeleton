package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFileID(t *testing.T) {
	id := NewFileID()
	require.NotEmpty(t, id)
	require.Len(t, id.String(), 36)
}

func TestParseFileID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid ID",
			input:   "019d65d6-de90-7200-b1cf-4f8745597e0a",
			wantErr: false,
		},
		{
			name:    "empty ID",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ParseFileID(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, ErrInvalidFileID, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, FileID(tt.input), id)
			}
		})
	}
}

func TestFileIDString(t *testing.T) {
	id := FileID("test-id-123")
	require.Equal(t, "test-id-123", id.String())
}

func TestNewUploadID(t *testing.T) {
	id := NewUploadID()
	require.NotEmpty(t, id)
	require.Len(t, id.String(), 36)
}

func TestParseUploadID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid ID",
			input:   "019d65d6-de90-7200-b1cf-4f8745597e0b",
			wantErr: false,
		},
		{
			name:    "empty ID",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ParseUploadID(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, ErrInvalidUploadID, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, UploadID(tt.input), id)
			}
		})
	}
}

func TestUploadIDString(t *testing.T) {
	id := UploadID("upload-id-456")
	require.Equal(t, "upload-id-456", id.String())
}

func TestNewProcessingID(t *testing.T) {
	id := NewProcessingID()
	require.NotEmpty(t, id)
	require.Len(t, id.String(), 36)
}

func TestParseProcessingID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid ID",
			input:   "019d65d6-de90-7200-b1cf-4f8745597e0c",
			wantErr: false,
		},
		{
			name:    "empty ID",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ParseProcessingID(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, ErrInvalidProcessingID, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, ProcessingID(tt.input), id)
			}
		})
	}
}

func TestProcessingIDString(t *testing.T) {
	id := ProcessingID("processing-id-789")
	require.Equal(t, "processing-id-789", id.String())
}

func TestFileIDUniqueness(t *testing.T) {
	id1 := NewFileID()
	id2 := NewFileID()
	require.NotEqual(t, id1, id2)
}

func TestUploadIDUniqueness(t *testing.T) {
	id1 := NewUploadID()
	id2 := NewUploadID()
	require.NotEqual(t, id1, id2)
}

func TestProcessingIDUniqueness(t *testing.T) {
	id1 := NewProcessingID()
	id2 := NewProcessingID()
	require.NotEqual(t, id1, id2)
}

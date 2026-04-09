package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewFileProcessing(t *testing.T) {
	fileID := NewFileID()

	t.Run("valid processing", func(t *testing.T) {
		options := ProcessingOptions{
			Width:   intPtr(800),
			Height:  intPtr(600),
			Quality: intPtr(85),
		}

		processing, err := NewFileProcessing(fileID, OperationResize, options)
		require.NoError(t, err)
		require.NotNil(t, processing)
		require.NotEmpty(t, processing.ID())
		require.Equal(t, fileID, processing.FileID())
		require.Equal(t, OperationResize, processing.Operation())
		require.Equal(t, options, processing.Options())
		require.Equal(t, ProcessingPending, processing.Status())
		require.False(t, processing.CreatedAt().IsZero())
	})

	t.Run("different operations", func(t *testing.T) {
		operations := []ProcessingOperation{
			OperationResize,
			OperationCrop,
			OperationCompress,
			OperationConvert,
			OperationThumbnail,
			OperationWatermark,
		}

		for _, op := range operations {
			processing, err := NewFileProcessing(fileID, op, ProcessingOptions{})
			require.NoError(t, err)
			require.Equal(t, op, processing.Operation())
		}
	})

	t.Run("invalid operation", func(t *testing.T) {
		_, err := NewFileProcessing(fileID, ProcessingOperation("invalid"), ProcessingOptions{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid processing operation")
	})
}

func TestFileProcessingStart(t *testing.T) {
	fileID := NewFileID()
	processing, _ := NewFileProcessing(fileID, OperationResize, ProcessingOptions{})

	t.Run("from pending", func(t *testing.T) {
		err := processing.Start()
		require.NoError(t, err)
		require.Equal(t, ProcessingRunning, processing.Status())
		require.NotNil(t, processing.StartedAt())
	})

	t.Run("from wrong status", func(t *testing.T) {
		processing2, _ := NewFileProcessing(fileID, OperationResize, ProcessingOptions{})
		_ = processing2.Start()
		_ = processing2.Complete(NewFileID())
		err := processing2.Start()
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot start processing")
	})
}

func TestFileProcessingComplete(t *testing.T) {
	fileID := NewFileID()
	resultFileID := NewFileID()

	t.Run("success", func(t *testing.T) {
		processing, _ := NewFileProcessing(fileID, OperationResize, ProcessingOptions{})
		_ = processing.Start()

		err := processing.Complete(resultFileID)
		require.NoError(t, err)
		require.Equal(t, ProcessingCompleted, processing.Status())
		require.NotNil(t, processing.ResultFileID())
		require.Equal(t, resultFileID, *processing.ResultFileID())
		require.NotNil(t, processing.CompletedAt())
		require.True(t, processing.IsCompleted())
	})

	t.Run("from wrong status", func(t *testing.T) {
		processing, _ := NewFileProcessing(fileID, OperationResize, ProcessingOptions{})

		err := processing.Complete(resultFileID)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot complete processing")
	})
}

func TestFileProcessingFail(t *testing.T) {
	fileID := NewFileID()

	t.Run("success", func(t *testing.T) {
		processing, _ := NewFileProcessing(fileID, OperationResize, ProcessingOptions{})
		_ = processing.Start()

		err := processing.Fail("Processing failed")
		require.NoError(t, err)
		require.Equal(t, ProcessingFailed, processing.Status())
		require.NotNil(t, processing.Error())
		require.Equal(t, "Processing failed", *processing.Error())
		require.NotNil(t, processing.CompletedAt())
		require.True(t, processing.IsFailed())
	})

	t.Run("from wrong status", func(t *testing.T) {
		processing, _ := NewFileProcessing(fileID, OperationResize, ProcessingOptions{})

		err := processing.Fail("Processing failed")
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot fail processing")
	})
}

func TestFileProcessingStatusMethods(t *testing.T) {
	fileID := NewFileID()

	t.Run("pending", func(t *testing.T) {
		processing, _ := NewFileProcessing(fileID, OperationResize, ProcessingOptions{})
		require.True(t, processing.IsPending())
		require.False(t, processing.IsRunning())
		require.False(t, processing.IsCompleted())
		require.False(t, processing.IsFailed())
	})

	t.Run("running", func(t *testing.T) {
		processing, _ := NewFileProcessing(fileID, OperationResize, ProcessingOptions{})
		_ = processing.Start()
		require.False(t, processing.IsPending())
		require.True(t, processing.IsRunning())
		require.False(t, processing.IsCompleted())
		require.False(t, processing.IsFailed())
	})

	t.Run("completed", func(t *testing.T) {
		processing, _ := NewFileProcessing(fileID, OperationResize, ProcessingOptions{})
		_ = processing.Start()
		_ = processing.Complete(NewFileID())
		require.False(t, processing.IsPending())
		require.False(t, processing.IsRunning())
		require.True(t, processing.IsCompleted())
		require.False(t, processing.IsFailed())
	})

	t.Run("failed", func(t *testing.T) {
		processing, _ := NewFileProcessing(fileID, OperationResize, ProcessingOptions{})
		_ = processing.Start()
		_ = processing.Fail("error")
		require.False(t, processing.IsPending())
		require.False(t, processing.IsRunning())
		require.False(t, processing.IsCompleted())
		require.True(t, processing.IsFailed())
	})
}

func TestFileProcessingDuration(t *testing.T) {
	fileID := NewFileID()

	t.Run("not started", func(t *testing.T) {
		processing, _ := NewFileProcessing(fileID, OperationResize, ProcessingOptions{})
		require.Nil(t, processing.Duration())
	})

	t.Run("started but not completed", func(t *testing.T) {
		processing, _ := NewFileProcessing(fileID, OperationResize, ProcessingOptions{})
		_ = processing.Start()
		require.Nil(t, processing.Duration())
	})

	t.Run("completed", func(t *testing.T) {
		processing, _ := NewFileProcessing(fileID, OperationResize, ProcessingOptions{})
		_ = processing.Start()
		time.Sleep(10 * time.Millisecond)
		_ = processing.Complete(NewFileID())
		duration := processing.Duration()
		require.NotNil(t, duration)
		require.True(t, *duration >= 10*time.Millisecond)
	})

	t.Run("failed", func(t *testing.T) {
		processing, _ := NewFileProcessing(fileID, OperationResize, ProcessingOptions{})
		_ = processing.Start()
		time.Sleep(10 * time.Millisecond)
		_ = processing.Fail("error")
		duration := processing.Duration()
		require.NotNil(t, duration)
		require.True(t, *duration >= 10*time.Millisecond)
	})
}

func TestReconstituteFileProcessing(t *testing.T) {
	fileID := NewFileID()
	resultFileID := NewFileID()
	processingID := NewProcessingID()
	now := time.Now()
	startedAt := now.Add(-1 * time.Minute)
	completedAt := now

	options := ProcessingOptions{
		Width:   intPtr(800),
		Height:  intPtr(600),
		Quality: intPtr(85),
	}

	processing := ReconstituteFileProcessing(
		processingID,
		fileID,
		OperationResize,
		options,
		ProcessingCompleted,
		&resultFileID,
		nil,
		&startedAt,
		&completedAt,
		now,
	)

	require.Equal(t, processingID, processing.ID())
	require.Equal(t, fileID, processing.FileID())
	require.Equal(t, OperationResize, processing.Operation())
	require.Equal(t, options, processing.Options())
	require.Equal(t, ProcessingCompleted, processing.Status())
	require.Equal(t, &resultFileID, processing.ResultFileID())
	require.Equal(t, &startedAt, processing.StartedAt())
	require.Equal(t, &completedAt, processing.CompletedAt())
}

func TestReconstituteFileProcessingWithError(t *testing.T) {
	fileID := NewFileID()
	errorMsg := "processing error"
	now := time.Now()
	startedAt := now.Add(-1 * time.Minute)
	completedAt := now

	processing := ReconstituteFileProcessing(
		NewProcessingID(),
		fileID,
		OperationResize,
		ProcessingOptions{},
		ProcessingFailed,
		nil,
		&errorMsg,
		&startedAt,
		&completedAt,
		now,
	)

	require.Equal(t, ProcessingFailed, processing.Status())
	require.Nil(t, processing.ResultFileID())
	require.Equal(t, &errorMsg, processing.Error())
	require.True(t, processing.IsFailed())
}

func TestProcessingOperationString(t *testing.T) {
	require.Equal(t, "resize", OperationResize.String())
	require.Equal(t, "crop", OperationCrop.String())
	require.Equal(t, "compress", OperationCompress.String())
	require.Equal(t, "convert", OperationConvert.String())
	require.Equal(t, "thumbnail", OperationThumbnail.String())
	require.Equal(t, "watermark", OperationWatermark.String())
}

func TestProcessingStatusString(t *testing.T) {
	require.Equal(t, "pending", ProcessingPending.String())
	require.Equal(t, "running", ProcessingRunning.String())
	require.Equal(t, "completed", ProcessingCompleted.String())
	require.Equal(t, "failed", ProcessingFailed.String())
}

func TestProcessingOptionsWithCustomFields(t *testing.T) {
	fileID := NewFileID()
	customOptions := ProcessingOptions{
		Width:  intPtr(1920),
		Height: intPtr(1080),
		Format: strPtr("webp"),
		Custom: map[string]string{"aspect_ratio": "16:9"},
	}

	processing, err := NewFileProcessing(fileID, OperationConvert, customOptions)
	require.NoError(t, err)
	require.Equal(t, customOptions, processing.Options())
	require.Equal(t, 1920, *processing.Options().Width)
	require.Equal(t, 1080, *processing.Options().Height)
	require.Equal(t, "webp", *processing.Options().Format)
	require.Equal(t, "16:9", processing.Options().Custom["aspect_ratio"])
}

func TestIsValidOperation(t *testing.T) {
	tests := []struct {
		name      string
		operation ProcessingOperation
		valid     bool
	}{
		{name: "resize", operation: OperationResize, valid: true},
		{name: "crop", operation: OperationCrop, valid: true},
		{name: "compress", operation: OperationCompress, valid: true},
		{name: "convert", operation: OperationConvert, valid: true},
		{name: "thumbnail", operation: OperationThumbnail, valid: true},
		{name: "watermark", operation: OperationWatermark, valid: true},
		{name: "invalid", operation: ProcessingOperation("invalid"), valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.valid, isValidOperation(tt.operation))
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func strPtr(s string) *string {
	return &s
}

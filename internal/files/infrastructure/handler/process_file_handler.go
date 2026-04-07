package handler

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/files/domain"
	"github.com/basilex/skeleton/internal/files/infrastructure/processing"
	"github.com/basilex/skeleton/internal/files/infrastructure/storage"
	tasksdomain "github.com/basilex/skeleton/internal/tasks/domain"
)

// ProcessFileHandler handles file processing tasks from the Tasks context.
type ProcessFileHandler struct {
	processingRepo domain.ProcessingRepository
	fileRepo       domain.FileRepository
	storage        storage.StorageProvider
	imageProcessor processing.ImageProcessor
}

// NewProcessFileHandler creates a new process file handler.
func NewProcessFileHandler(
	processingRepo domain.ProcessingRepository,
	fileRepo domain.FileRepository,
	storage storage.StorageProvider,
	imageProcessor processing.ImageProcessor,
) *ProcessFileHandler {
	return &ProcessFileHandler{
		processingRepo: processingRepo,
		fileRepo:       fileRepo,
		storage:        storage,
		imageProcessor: imageProcessor,
	}
}

// Handle processes a file processing task.
func (h *ProcessFileHandler) Handle(ctx context.Context, task *tasksdomain.Task) error {
	processingID, ok := task.Payload()["processing_id"].(string)
	if !ok {
		return fmt.Errorf("missing processing_id")
	}

	processing, err := h.processingRepo.GetByID(ctx, domain.ProcessingID(processingID))
	if err != nil {
		return fmt.Errorf("get processing: %w", err)
	}

	if err := processing.Start(); err != nil {
		return err
	}
	if err := h.processingRepo.Update(ctx, processing); err != nil {
		return err
	}

	file, err := h.fileRepo.GetByID(ctx, processing.FileID())
	if err != nil {
		_ = processing.Fail(fmt.Sprintf("get file: %v", err))
		_ = h.processingRepo.Update(ctx, processing)
		return err
	}

	data, err := h.storage.DownloadToBytes(ctx, file.Path())
	if err != nil {
		_ = processing.Fail(fmt.Sprintf("download: %v", err))
		_ = h.processingRepo.Update(ctx, processing)
		return err
	}

	resultData, err := h.processImage(ctx, file, data, processing)
	if err != nil {
		_ = processing.Fail(err.Error())
		_ = h.processingRepo.Update(ctx, processing)
		return err
	}

	resultPath := fmt.Sprintf("%s_%s", file.Path(), processing.Operation())
	if err := h.storage.UploadFromBytes(ctx, resultPath, resultData, file.MimeType()); err != nil {
		_ = processing.Fail(fmt.Sprintf("upload: %v", err))
		_ = h.processingRepo.Update(ctx, processing)
		return err
	}

	resultFile, err := domain.NewFile(file.OwnerID(), file.Filename()+"_"+string(processing.Operation()), file.MimeType(), int64(len(resultData)), file.StorageProvider(), file.AccessLevel())
	if err != nil {
		_ = processing.Fail(fmt.Sprintf("create file: %v", err))
		_ = h.processingRepo.Update(ctx, processing)
		return err
	}

	resultFile.SetPath(resultPath)
	resultFile.SetChecksum(resultData)
	resultFile.SetProcessed()

	if err := h.fileRepo.Create(ctx, resultFile); err != nil {
		_ = processing.Fail(fmt.Sprintf("save file: %v", err))
		_ = h.processingRepo.Update(ctx, processing)
		return err
	}

	processing.Complete(resultFile.ID())
	return h.processingRepo.Update(ctx, processing)
}

func (h *ProcessFileHandler) processImage(ctx context.Context, file *domain.File, data []byte, proc *domain.FileProcessing) ([]byte, error) {
	opts := proc.Options()
	switch proc.Operation() {
	case domain.OperationResize:
		w, h1 := 0, 0
		if opts.Width != nil {
			w = *opts.Width
		}
		if opts.Height != nil {
			h1 = *opts.Height
		}
		return h.imageProcessor.Resize(ctx, data, w, h1)
	case domain.OperationThumbnail:
		w, h1 := 150, 150
		if opts.Width != nil {
			w = *opts.Width
		}
		if opts.Height != nil {
			h1 = *opts.Height
		}
		return h.imageProcessor.GenerateThumbnail(ctx, data, w, h1)
	case domain.OperationCompress:
		q := 85
		if opts.Quality != nil {
			q = *opts.Quality
		}
		return h.imageProcessor.Compress(ctx, data, q)
	case domain.OperationConvert:
		f := "jpeg"
		if opts.Format != nil {
			f = *opts.Format
		}
		return h.imageProcessor.Convert(ctx, data, f)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", proc.Operation())
	}
}

// CanHandle returns true if this handler can handle the task type.
func (h *ProcessFileHandler) CanHandle(taskType tasksdomain.TaskType) bool {
	switch taskType {
	case "process_file_resize", "process_file_crop", "process_file_compress",
		"process_file_convert", "process_file_thumbnail", "process_file_watermark":
		return true
	}
	return false
}

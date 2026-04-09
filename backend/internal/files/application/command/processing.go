package command

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/files/domain"
	tasksdomain "github.com/basilex/skeleton/internal/tasks/domain"
)

type requestProcessingHandler struct {
	processingRepo domain.ProcessingRepository
	fileRepo       domain.FileRepository
	taskRepo       tasksdomain.TaskRepository
}

func NewRequestProcessingHandler(
	processingRepo domain.ProcessingRepository,
	fileRepo domain.FileRepository,
	taskRepo tasksdomain.TaskRepository,
) RequestProcessingHandler {
	h := &requestProcessingHandler{
		processingRepo: processingRepo,
		fileRepo:       fileRepo,
		taskRepo:       taskRepo,
	}
	return h.Request
}

func (h *requestProcessingHandler) Request(ctx context.Context, cmd RequestProcessingCommand) (*RequestProcessingResult, error) {
	if err := cmd.validate(); err != nil {
		return nil, err
	}

	fileID, err := ParseFileID(cmd.FileID)
	if err != nil {
		return nil, err
	}

	file, err := h.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("get file: %w", err)
	}

	if !file.IsImage() {
		return nil, domain.NewValidationError("fileID", "file must be an image for processing")
	}

	processing, err := domain.NewFileProcessing(fileID, cmd.Operation, cmd.Options)
	if err != nil {
		return nil, fmt.Errorf("create processing: %w", err)
	}

	if err := h.processingRepo.Create(ctx, processing); err != nil {
		return nil, fmt.Errorf("save processing: %w", err)
	}

	taskType := tasksdomain.TaskType(fmt.Sprintf("process_file_%s", cmd.Operation))
	taskPayload := tasksdomain.TaskPayload{
		"processing_id": processing.ID().String(),
		"file_id":       fileID.String(),
		"operation":     string(cmd.Operation),
	}

	if cmd.Options.Width != nil {
		taskPayload["width"] = *cmd.Options.Width
	}
	if cmd.Options.Height != nil {
		taskPayload["height"] = *cmd.Options.Height
	}
	if cmd.Options.Quality != nil {
		taskPayload["quality"] = *cmd.Options.Quality
	}
	if cmd.Options.Format != nil {
		taskPayload["format"] = *cmd.Options.Format
	}

	task, err := tasksdomain.NewTask(taskType, taskPayload)
	if err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}

	if err := h.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("save task: %w", err)
	}

	return &RequestProcessingResult{
		ProcessingID: processing.ID().String(),
		Status:       string(processing.Status()),
	}, nil
}

// Helper functions for processing lifecycle
func StartProcessing(ctx context.Context, processingRepo domain.ProcessingRepository, processingID string) error {
	pid, err := domain.ParseProcessingID(processingID)
	if err != nil {
		return err
	}

	processing, err := processingRepo.GetByID(ctx, pid)
	if err != nil {
		return fmt.Errorf("get processing: %w", err)
	}

	if err := processing.Start(); err != nil {
		return err
	}

	return processingRepo.Update(ctx, processing)
}

func CompleteProcessing(ctx context.Context, processingRepo domain.ProcessingRepository, processingID string, resultFileID string) error {
	pid, err := domain.ParseProcessingID(processingID)
	if err != nil {
		return err
	}

	rfid, err := domain.ParseFileID(resultFileID)
	if err != nil {
		return err
	}

	processing, err := processingRepo.GetByID(ctx, pid)
	if err != nil {
		return fmt.Errorf("get processing: %w", err)
	}

	if err := processing.Complete(rfid); err != nil {
		return err
	}

	return processingRepo.Update(ctx, processing)
}

func FailProcessing(ctx context.Context, processingRepo domain.ProcessingRepository, processingID string, errorMessage string) error {
	pid, err := domain.ParseProcessingID(processingID)
	if err != nil {
		return err
	}

	processing, err := processingRepo.GetByID(ctx, pid)
	if err != nil {
		return fmt.Errorf("get processing: %w", err)
	}

	if err := processing.Fail(errorMessage); err != nil {
		return err
	}

	return processingRepo.Update(ctx, processing)
}

func RetryDelay(attempts int) time.Duration {
	delays := []time.Duration{
		1 * time.Second,
		5 * time.Second,
		15 * time.Second,
		1 * time.Minute,
		5 * time.Minute,
		15 * time.Minute,
		1 * time.Hour,
	}

	if attempts >= len(delays) {
		return delays[len(delays)-1]
	}

	return delays[attempts]
}

package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/files/domain"
)

type getFileHandler struct {
	fileRepo domain.FileRepository
}

func NewGetFileHandler(fileRepo domain.FileRepository) GetFileHandler {
	h := &getFileHandler{fileRepo: fileRepo}
	return h.Get
}

func (h *getFileHandler) Get(ctx context.Context, query GetFileQuery) (*FileDTO, error) {
	fileID, err := domain.ParseFileID(query.FileID)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID: %w", err)
	}

	file, err := h.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("get file: %w", err)
	}

	return ToFileDTO(file), nil
}

type listFilesHandler struct {
	fileRepo domain.FileRepository
}

func NewListFilesHandler(fileRepo domain.FileRepository) ListFilesHandler {
	h := &listFilesHandler{fileRepo: fileRepo}
	return h.List
}

func (h *listFilesHandler) List(ctx context.Context, query ListFilesQuery) (*ListFilesResult, error) {
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	filter := &domain.FileFilter{
		OwnerID:      query.OwnerID,
		MimeType:     query.MimeType,
		AccessLevel:  query.AccessLevel,
		UploadedFrom: query.UploadedFrom,
		UploadedTo:   query.UploadedTo,
	}

	files, err := h.fileRepo.List(ctx, filter, query.Limit+1, 0)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}

	hasMore := len(files) > query.Limit
	if hasMore {
		files = files[:query.Limit]
	}

	var nextCursor *string
	if hasMore && len(files) > 0 {
		cursor := files[len(files)-1].ID().String()
		nextCursor = &cursor
	}

	dtos := make([]FileDTO, len(files))
	for i, file := range files {
		dtos[i] = *ToFileDTO(file)
	}

	return &ListFilesResult{
		Items:      dtos,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

type listProcessingsHandler struct {
	processingRepo domain.ProcessingRepository
}

func NewListProcessingsHandler(processingRepo domain.ProcessingRepository) ListProcessingsHandler {
	h := &listProcessingsHandler{processingRepo: processingRepo}
	return h.List
}

func (h *listProcessingsHandler) List(ctx context.Context, query ListProcessingsQuery) (*ListProcessingsResult, error) {
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	fileID, err := domain.ParseFileID(query.FileID)
	if err != nil {
		return nil, fmt.Errorf("invalid file ID: %w", err)
	}

	filter := &domain.ProcessingFilter{
		FileID: &fileID,
		Status: query.Status,
	}

	processings, err := h.processingRepo.List(ctx, filter, query.Limit+1, 0)
	if err != nil {
		return nil, fmt.Errorf("list processings: %w", err)
	}

	hasMore := len(processings) > query.Limit
	if hasMore {
		processings = processings[:query.Limit]
	}

	var nextCursor *string
	if hasMore && len(processings) > 0 {
		cursor := processings[len(processings)-1].ID().String()
		nextCursor = &cursor
	}

	dtos := make([]ProcessingDTO, len(processings))
	for i, processing := range processings {
		dtos[i] = *ToProcessingDTO(processing)
	}

	return &ListProcessingsResult{
		Items:      dtos,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

type getProcessingStatusHandler struct {
	processingRepo domain.ProcessingRepository
}

func NewGetProcessingStatusHandler(processingRepo domain.ProcessingRepository) GetProcessingStatusHandler {
	h := &getProcessingStatusHandler{processingRepo: processingRepo}
	return h.Get
}

func (h *getProcessingStatusHandler) Get(ctx context.Context, query GetProcessingStatusQuery) (*ProcessingDTO, error) {
	processingID, err := domain.ParseProcessingID(query.ProcessingID)
	if err != nil {
		return nil, fmt.Errorf("invalid processing ID: %w", err)
	}

	processing, err := h.processingRepo.GetByID(ctx, processingID)
	if err != nil {
		return nil, fmt.Errorf("get processing: %w", err)
	}

	return ToProcessingDTO(processing), nil
}

package http

import (
	"net/http"

	"github.com/basilex/skeleton/internal/files/application/command"
	"github.com/basilex/skeleton/internal/files/application/query"
	"github.com/basilex/skeleton/internal/files/domain"
	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for files.
type Handler struct {
	uploadFile          command.UploadFileHandler
	deleteFile          command.DeleteFileHandler
	requestUploadURL    command.RequestUploadURLHandler
	confirmUpload       command.ConfirmUploadHandler
	requestProcessing   command.RequestProcessingHandler
	getFile             query.GetFileHandler
	listFiles           query.ListFilesHandler
	getProcessingStatus query.GetProcessingStatusHandler
	listProcessings     query.ListProcessingsHandler
}

// NewHandler creates a new files handler.
func NewHandler(
	uploadFile command.UploadFileHandler,
	deleteFile command.DeleteFileHandler,
	requestUploadURL command.RequestUploadURLHandler,
	confirmUpload command.ConfirmUploadHandler,
	requestProcessing command.RequestProcessingHandler,
	getFile query.GetFileHandler,
	listFiles query.ListFilesHandler,
	getProcessingStatus query.GetProcessingStatusHandler,
	listProcessings query.ListProcessingsHandler,
) *Handler {
	return &Handler{
		uploadFile:          uploadFile,
		deleteFile:          deleteFile,
		requestUploadURL:    requestUploadURL,
		confirmUpload:       confirmUpload,
		requestProcessing:   requestProcessing,
		getFile:             getFile,
		listFiles:           listFiles,
		getProcessingStatus: getProcessingStatus,
		listProcessings:     listProcessings,
	}
}

// UploadFile godoc
// @Summary Upload a file
// @Description Upload a file directly (for small files <5MB)
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Param owner_id formData string false "Owner ID"
// @Param access_level formData string true "Access level (public/private/restricted)"
// @Success 201 {object} UploadFileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files [post]
func (h *Handler) UploadFile(c *gin.Context) {
	maxSize := int64(5 << 20) // 5 MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		if err.Error() == "http: request body too large" {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{
				Error:   "file_too_large",
				Message: "File size exceeds 5MB limit. Use presigned URL upload for larger files.",
			})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to get file from request: " + err.Error(),
		})
		return
	}
	defer file.Close()

	ownerID := c.PostForm("owner_id")
	accessLevel := c.PostForm("access_level")
	if accessLevel == "" {
		accessLevel = "private"
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	cmd := command.UploadFileCommand{
		Filename:        header.Filename,
		Content:         file,
		MimeType:        contentType,
		Size:            header.Size,
		StorageProvider: domain.StorageLocal,
		AccessLevel:     domain.AccessLevel(accessLevel),
	}

	if ownerID != "" {
		cmd.OwnerID = &ownerID
	}

	result, err := h.uploadFile(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "upload_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, UploadFileResponse{
		FileID:      result.FileID,
		StoredName:  result.StoredName,
		StoragePath: result.StoragePath,
		Checksum:    result.Checksum,
		UploadedAt:  result.UploadedAt,
	})
}

// GetFile godoc
// @Summary Get file metadata
// @Description Get file metadata by ID
// @Tags files
// @Produce json
// @Param id path string true "File ID"
// @Success 200 {object} FileResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files/{id} [get]
func (h *Handler) GetFile(c *gin.Context) {
	fileID := c.Param("id")

	file, err := h.getFile(c.Request.Context(), query.GetFileQuery{FileID: fileID})
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, file)
}

// ListFiles godoc
// @Summary List files
// @Description List files with filtering and pagination
// @Tags files
// @Produce json
// @Param owner_id query string false "Filter by owner ID"
// @Param mime_type query string false "Filter by MIME type"
// @Param access_level query string false "Filter by access level"
// @Param limit query int false "Limit" default(20)
// @Param cursor query string false "Cursor for pagination"
// @Success 200 {object} ListFilesResponse
// @Failure 500 {object} ErrorResponse
// @Router /files [get]
func (h *Handler) ListFiles(c *gin.Context) {
	var req ListFilesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	var accessLevel *domain.AccessLevel
	if req.AccessLevel != nil {
		al := domain.AccessLevel(*req.AccessLevel)
		accessLevel = &al
	}

	result, err := h.listFiles(c.Request.Context(), query.ListFilesQuery{
		OwnerID:     req.OwnerID,
		MimeType:    req.MimeType,
		AccessLevel: accessLevel,
		Limit:       req.Limit,
		Cursor:      req.Cursor,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// DeleteFile godoc
// @Summary Delete a file
// @Description Delete a file by ID
// @Tags files
// @Param id path string true "File ID"
// @Success 204 "No Content"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files/{id} [delete]
func (h *Handler) DeleteFile(c *gin.Context) {
	fileID := c.Param("id")

	if err := h.deleteFile(c.Request.Context(), command.DeleteFileCommand{FileID: fileID}); err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// RequestUploadURL godoc
// @Summary Request presigned upload URL
// @Description Request a presigned URL for large file uploads
// @Tags files
// @Accept json
// @Produce json
// @Param request body UploadURLRequest true "Upload URL request"
// @Success 200 {object} UploadURLResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files/upload-url [post]
func (h *Handler) RequestUploadURL(c *gin.Context) {
	var req UploadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	result, err := h.requestUploadURL(c.Request.Context(), command.RequestUploadURLCommand{
		OwnerID:         req.OwnerID,
		Filename:        req.Filename,
		MimeType:        req.MimeType,
		Size:            req.Size,
		StorageProvider: domain.StorageProvider(req.StorageProvider),
		AccessLevel:     domain.AccessLevel(req.AccessLevel),
		TTL:             req.TTL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ConfirmUpload godoc
// @Summary Confirm upload
// @Description Confirm that a file has been uploaded using presigned URL
// @Tags files
// @Accept json
// @Produce json
// @Param request body ConfirmUploadRequest true "Confirm upload request"
// @Success 200 {object} ConfirmUploadResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files/confirm [post]
func (h *Handler) ConfirmUpload(c *gin.Context) {
	var req ConfirmUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	result, err := h.confirmUpload(c.Request.Context(), command.ConfirmUploadCommand{
		UploadID: req.UploadID,
		Checksum: req.Checksum,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RequestProcessing godoc
// @Summary Request file processing
// @Description Request image processing (resize, crop, compress, etc.)
// @Tags files
// @Accept json
// @Produce json
// @Param id path string true "File ID"
// @Param request body ProcessFileRequest true "Processing request"
// @Success 200 {object} ProcessFileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files/{id}/process [post]
func (h *Handler) RequestProcessing(c *gin.Context) {
	fileID := c.Param("id")

	var req ProcessFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Parse options
	var options domain.ProcessingOptions
	if req.Options != nil {
		if w, ok := req.Options["width"].(float64); ok {
			width := int(w)
			options.Width = &width
		}
		if h, ok := req.Options["height"].(float64); ok {
			height := int(h)
			options.Height = &height
		}
		if q, ok := req.Options["quality"].(float64); ok {
			quality := int(q)
			options.Quality = &quality
		}
		if f, ok := req.Options["format"].(string); ok {
			options.Format = &f
		}
	}

	result, err := h.requestProcessing(c.Request.Context(), command.RequestProcessingCommand{
		FileID:    fileID,
		Operation: domain.ProcessingOperation(req.Operation),
		Options:   options,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetProcessingStatus godoc
// @Summary Get processing status
// @Description Get file processing status by ID
// @Tags files
// @Produce json
// @Param id path string true "Processing ID"
// @Success 200 {object} ProcessingResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files/processing/{id} [get]
func (h *Handler) GetProcessingStatus(c *gin.Context) {
	processingID := c.Param("id")

	processing, err := h.getProcessingStatus(c.Request.Context(), query.GetProcessingStatusQuery{
		ProcessingID: processingID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "not_found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, processing)
}

// ListProcessings godoc
// @Summary List processings for a file
// @Description List all processing operations for a file
// @Tags files
// @Produce json
// @Param file_id query string true "File ID"
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit" default(20)
// @Success 200 {object} ListProcessingsResponse
// @Failure 500 {object} ErrorResponse
// @Router /files/processing [get]
func (h *Handler) ListProcessings(c *gin.Context) {
	fileID := c.Query("file_id")
	if fileID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "file_id is required",
		})
		return
	}

	var req ListProcessingsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}

	var status *domain.ProcessingStatus
	if req.Status != nil {
		s := domain.ProcessingStatus(*req.Status)
		status = &s
	}

	result, err := h.listProcessings(c.Request.Context(), query.ListProcessingsQuery{
		FileID: fileID,
		Status: status,
		Limit:  req.Limit,
		Cursor: req.Cursor,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

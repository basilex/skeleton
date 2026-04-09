package http

import (
	"errors"
	"fmt"
	"net/http"

	documentsCommand "github.com/basilex/skeleton/internal/documents/application/command"
	documentsQuery "github.com/basilex/skeleton/internal/documents/application/query"
	documentsDomain "github.com/basilex/skeleton/internal/documents/domain"
	"github.com/basilex/skeleton/pkg/apierror"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	createDocument   *documentsCommand.CreateDocumentHandler
	generateDocument *documentsCommand.GenerateDocumentHandler
	addSignature     *documentsCommand.AddSignatureHandler
	signDocument     *documentsCommand.SignDocumentHandler
	getDocument      *documentsQuery.GetDocumentHandler
	listDocuments    *documentsQuery.ListDocumentsHandler
	getTemplate      *documentsQuery.GetTemplateHandler
}

func NewHandler(
	createDocument *documentsCommand.CreateDocumentHandler,
	generateDocument *documentsCommand.GenerateDocumentHandler,
	addSignature *documentsCommand.AddSignatureHandler,
	signDocument *documentsCommand.SignDocumentHandler,
	getDocument *documentsQuery.GetDocumentHandler,
	listDocuments *documentsQuery.ListDocumentsHandler,
	getTemplate *documentsQuery.GetTemplateHandler,
) *Handler {
	return &Handler{
		createDocument:   createDocument,
		generateDocument: generateDocument,
		addSignature:     addSignature,
		signDocument:     signDocument,
		getDocument:      getDocument,
		listDocuments:    listDocuments,
		getTemplate:      getTemplate,
	}
}

// CreateDocument godoc
// @Summary Create a new document
// @Description Creates a new document in draft status
// @Tags documents
// @Accept json
// @Produce json
// @Param request body CreateDocumentRequest true "Document data"
// @Success 201 {object} map[string]string "Document created"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/documents [post]
func (h *Handler) CreateDocument(c *gin.Context) {
	var req CreateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.createDocument.Handle(c.Request.Context(), documentsCommand.CreateDocumentCommand{
		DocumentNumber: req.DocumentNumber,
		DocumentType:   req.DocumentType,
		ReferenceID:    req.ReferenceID,
	})
	if err != nil {
		handleDocumentsError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": result.DocumentID})
}

// GenerateDocument godoc
// @Summary Generate document from template
// @Description Generates a PDF document from a template
// @Tags documents
// @Accept json
// @Produce json
// @Param id path string true "Document ID"
// @Param request body GenerateDocumentRequest true "Template and data"
// @Success 200 {object} map[string]string "Document generated"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 404 {object} apierror.APIError "Document not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/documents/{id}/generate [post]
func (h *Handler) GenerateDocument(c *gin.Context) {
	documentID := c.Param("id")

	var req GenerateDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.generateDocument.Handle(c.Request.Context(), documentsCommand.GenerateDocumentCommand{
		DocumentID: documentID,
		TemplateID: req.TemplateID,
		Data:       req.Data,
	})
	if err != nil {
		handleDocumentsError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"file_id": result.FileID})
}

// AddSignature godoc
// @Summary Add signature to document
// @Description Adds a signature request to a document
// @Tags documents
// @Accept json
// @Produce json
// @Param id path string true "Document ID"
// @Param request body AddSignatureRequest true "Signature data"
// @Success 201 {object} map[string]string "Signature added"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 404 {object} apierror.APIError "Document not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/documents/{id}/signatures [post]
func (h *Handler) AddSignature(c *gin.Context) {
	documentID := c.Param("id")

	var req AddSignatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	result, err := h.addSignature.Handle(c.Request.Context(), documentsCommand.AddSignatureCommand{
		DocumentID: documentID,
		SignerName: req.SignerName,
		SignerRole: req.SignerRole,
	})
	if err != nil {
		handleDocumentsError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"signature_id": result.SignatureID})
}

// SignDocument godoc
// @Summary Sign document
// @Description Signs a document with digital signature
// @Tags documents
// @Accept json
// @Produce json
// @Param id path string true "Document ID"
// @Param request body SignDocumentRequest true "Signature data"
// @Success 200 {object} map[string]string "Document signed"
// @Failure 422 {object} apierror.APIError "Validation error"
// @Failure 404 {object} apierror.APIError "Document not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/documents/{id}/sign [post]
func (h *Handler) SignDocument(c *gin.Context) {
	documentID := c.Param("id")

	var req SignDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	err := h.signDocument.Handle(c.Request.Context(), documentsCommand.SignDocumentCommand{
		DocumentID:    documentID,
		SignatureID:   req.SignatureID,
		SignatureData: req.SignatureData,
	})
	if err != nil {
		handleDocumentsError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "document signed"})
}

// GetDocument godoc
// @Summary Get document by ID
// @Description Retrieves a document with its signatures by ID
// @Tags documents
// @Produce json
// @Param id path string true "Document ID"
// @Success 200 {object} DocumentResponse "Document details"
// @Failure 404 {object} apierror.APIError "Document not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/documents/{id} [get]
func (h *Handler) GetDocument(c *gin.Context) {
	documentID := c.Param("id")

	result, err := h.getDocument.Handle(c.Request.Context(), documentsQuery.GetDocumentQuery{
		DocumentID: documentID,
	})
	if err != nil {
		handleDocumentsError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ListDocuments godoc
// @Summary List documents
// @Description Lists documents with filtering and pagination
// @Tags documents
// @Produce json
// @Param document_type query string false "Filter by document type"
// @Param status query string false "Filter by status"
// @Param reference_id query string false "Filter by reference ID"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{} "List of documents"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/documents [get]
func (h *Handler) ListDocuments(c *gin.Context) {
	var query documentsQuery.ListDocumentsQuery

	if documentType := c.Query("document_type"); documentType != "" {
		query.DocumentType = &documentType
	}
	if status := c.Query("status"); status != "" {
		query.Status = &status
	}
	if referenceID := c.Query("reference_id"); referenceID != "" {
		query.ReferenceID = &referenceID
	}
	if cursor := c.Query("cursor"); cursor != "" {
		query.Cursor = cursor
	}
	if limit := c.Query("limit"); limit != "" {
		var l int
		if _, err := fmt.Sscanf(limit, "%d", &l); err == nil && l > 0 {
			query.Limit = l
		}
	}

	result, err := h.listDocuments.Handle(c.Request.Context(), query)
	if err != nil {
		handleDocumentsError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetTemplate godoc
// @Summary Get template by ID
// @Description Retrieves a document template by ID
// @Tags documents
// @Produce json
// @Param id path string true "Template ID"
// @Success 200 {object} TemplateResponse "Template details"
// @Failure 404 {object} apierror.APIError "Template not found"
// @Failure 500 {object} apierror.APIError "Internal error"
// @Router /api/v1/templates/{id} [get]
func (h *Handler) GetTemplate(c *gin.Context) {
	templateID := c.Param("id")

	result, err := h.getTemplate.Handle(c.Request.Context(), documentsQuery.GetTemplateQuery{
		TemplateID: templateID,
	})
	if err != nil {
		handleDocumentsError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func handleDocumentsError(c *gin.Context, err error) {
	if errors.Is(err, documentsDomain.ErrDocumentNotFound) {
		apierror.RespondError(c, apierror.NewNotFound("document not found", c.Request.URL.Path, getRequestID(c)))
		return
	}
	if errors.Is(err, documentsDomain.ErrTemplateNotFound) {
		apierror.RespondError(c, apierror.NewNotFound("template not found", c.Request.URL.Path, getRequestID(c)))
		return
	}
	if errors.Is(err, documentsDomain.ErrInvalidDocumentType) {
		apierror.RespondError(c, apierror.NewValidation(err.Error(), c.Request.URL.Path, getRequestID(c)))
		return
	}

	apierror.RespondError(c, apierror.NewInternal(err.Error(), c.Request.URL.Path, getRequestID(c)))
}

func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

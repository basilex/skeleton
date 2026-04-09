package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/documents/domain"
)

type GetDocumentHandler struct {
	documents domain.DocumentRepository
}

func NewGetDocumentHandler(documents domain.DocumentRepository) *GetDocumentHandler {
	return &GetDocumentHandler{documents: documents}
}

type GetDocumentQuery struct {
	DocumentID string
}

type DocumentDTO struct {
	ID             string            `json:"id"`
	DocumentNumber string            `json:"document_number"`
	DocumentType   string            `json:"document_type"`
	ReferenceID    string            `json:"reference_id"`
	FileID         *string           `json:"file_id"`
	Status         string            `json:"status"`
	Metadata       map[string]string `json:"metadata"`
	Signatures     []SignatureDTO    `json:"signatures"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
}

type SignatureDTO struct {
	ID            string  `json:"id"`
	SignerName    string  `json:"signer_name"`
	SignerRole    string  `json:"signer_role"`
	Status        string  `json:"status"`
	SignedAt      *string `json:"signed_at"`
	SignatureData string  `json:"signature_data,omitempty"`
}

func (h *GetDocumentHandler) Handle(ctx context.Context, query GetDocumentQuery) (*DocumentDTO, error) {
	docID, err := domain.ParseDocumentID(query.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("parse document ID: %w", err)
	}

	document, err := h.documents.FindByID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("find document: %w", err)
	}

	return toDocumentDTO(document), nil
}

func toDocumentDTO(document *domain.Document) *DocumentDTO {
	signatures := make([]SignatureDTO, 0, len(document.GetSignatures()))
	for _, sig := range document.GetSignatures() {
		var signedAt *string
		if sig.GetSignedAt() != nil {
			t := sig.GetSignedAt().Format("2006-01-02T15:04:05Z07:00")
			signedAt = &t
		}

		signatures = append(signatures, SignatureDTO{
			ID:            sig.GetID().String(),
			SignerName:    sig.GetSignerName(),
			SignerRole:    sig.GetSignerRole(),
			Status:        sig.GetStatus().String(),
			SignedAt:      signedAt,
			SignatureData: sig.GetSignatureData(),
		})
	}

	return &DocumentDTO{
		ID:             document.GetID().String(),
		DocumentNumber: document.GetDocumentNumber(),
		DocumentType:   document.GetDocumentType().String(),
		ReferenceID:    document.GetReferenceID(),
		FileID:         document.GetFileID(),
		Status:         document.GetStatus().String(),
		Metadata:       document.GetMetadata(),
		Signatures:     signatures,
		CreatedAt:      document.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      document.GetUpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}

type ListDocumentsHandler struct {
	documents domain.DocumentRepository
}

func NewListDocumentsHandler(documents domain.DocumentRepository) *ListDocumentsHandler {
	return &ListDocumentsHandler{documents: documents}
}

type ListDocumentsQuery struct {
	DocumentType  *string
	Status        *string
	ReferenceID   *string
	CreatedAfter  *string
	CreatedBefore *string
	Cursor        string
	Limit         int
}

func (h *ListDocumentsHandler) Handle(ctx context.Context, query ListDocumentsQuery) (interface{}, error) {
	var docType *domain.DocumentType
	if query.DocumentType != nil {
		dt := domain.DocumentType(*query.DocumentType)
		docType = &dt
	}

	var status *domain.DocumentStatus
	if query.Status != nil {
		s := domain.DocumentStatus(*query.Status)
		status = &s
	}

	filter := domain.DocumentFilter{
		DocumentType:  docType,
		Status:        status,
		ReferenceID:   query.ReferenceID,
		CreatedAfter:  query.CreatedAfter,
		CreatedBefore: query.CreatedBefore,
		Cursor:        query.Cursor,
		Limit:         query.Limit,
	}

	result, err := h.documents.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("find documents: %w", err)
	}

	dtos := make([]*DocumentDTO, 0, len(result.Items))
	for _, doc := range result.Items {
		dtos = append(dtos, toDocumentDTO(doc))
	}

	return map[string]interface{}{
		"items":       dtos,
		"next_cursor": result.NextCursor,
		"has_more":    result.HasMore,
	}, nil
}

type GetTemplateHandler struct {
	templates domain.TemplateRepository
}

func NewGetTemplateHandler(templates domain.TemplateRepository) *GetTemplateHandler {
	return &GetTemplateHandler{templates: templates}
}

type GetTemplateQuery struct {
	TemplateID string
}

type TemplateDTO struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	DocumentType string   `json:"document_type"`
	Content      string   `json:"content"`
	Variables    []string `json:"variables"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

func (h *GetTemplateHandler) Handle(ctx context.Context, query GetTemplateQuery) (*TemplateDTO, error) {
	templateID, err := domain.ParseTemplateID(query.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("parse template ID: %w", err)
	}

	template, err := h.templates.FindByID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("find template: %w", err)
	}

	return &TemplateDTO{
		ID:           template.GetID().String(),
		Name:         template.GetName(),
		DocumentType: template.GetDocumentType().String(),
		Content:      template.GetContent(),
		Variables:    template.GetVariables(),
		CreatedAt:    template.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    template.GetUpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

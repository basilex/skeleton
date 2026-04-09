package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/documents/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type CreateDocumentHandler struct {
	documents domain.DocumentRepository
	bus       eventbus.Bus
}

func NewCreateDocumentHandler(documents domain.DocumentRepository, bus eventbus.Bus) *CreateDocumentHandler {
	return &CreateDocumentHandler{
		documents: documents,
		bus:       bus,
	}
}

type CreateDocumentCommand struct {
	DocumentNumber string
	DocumentType   string
	ReferenceID    string
}

type CreateDocumentResult struct {
	DocumentID string
}

func (h *CreateDocumentHandler) Handle(ctx context.Context, cmd CreateDocumentCommand) (*CreateDocumentResult, error) {
	docType := domain.DocumentType(cmd.DocumentType)
	if !docType.IsValid() {
		return nil, domain.ErrInvalidDocumentType
	}

	document, err := domain.NewDocument(cmd.DocumentNumber, docType, cmd.ReferenceID)
	if err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}

	if err := h.documents.Save(ctx, document); err != nil {
		return nil, fmt.Errorf("save document: %w", err)
	}

	events := document.PullEvents()
	for _, event := range events {
		_ = h.bus.Publish(ctx, event)
	}

	return &CreateDocumentResult{
		DocumentID: document.GetID().String(),
	}, nil
}

type GenerateDocumentHandler struct {
	documents domain.DocumentRepository
	templates domain.TemplateRepository
	bus       eventbus.Bus
}

func NewGenerateDocumentHandler(documents domain.DocumentRepository, templates domain.TemplateRepository, bus eventbus.Bus) *GenerateDocumentHandler {
	return &GenerateDocumentHandler{
		documents: documents,
		templates: templates,
		bus:       bus,
	}
}

type GenerateDocumentCommand struct {
	DocumentID string
	TemplateID string
	Data       map[string]string
}

type GenerateDocumentResult struct {
	FileID string
}

func (h *GenerateDocumentHandler) Handle(ctx context.Context, cmd GenerateDocumentCommand) (*GenerateDocumentResult, error) {
	docID, err := domain.ParseDocumentID(cmd.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("parse document ID: %w", err)
	}

	document, err := h.documents.FindByID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("find document: %w", err)
	}

	templateID, err := domain.ParseTemplateID(cmd.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("parse template ID: %w", err)
	}

	// TODO: Implement PDF generation from template
	// This would integrate with the Files context to store the generated PDF
	_ = templateID // Template found, would be used for generation
	_ = cmd.Data   // Data would be merged with template

	// For now, generate a placeholder file ID
	fileID := "generated-file-id"

	if err := document.MarkGenerated(fileID); err != nil {
		return nil, fmt.Errorf("mark document generated: %w", err)
	}

	if err := h.documents.Save(ctx, document); err != nil {
		return nil, fmt.Errorf("save document: %w", err)
	}

	events := document.PullEvents()
	for _, event := range events {
		_ = h.bus.Publish(ctx, event)
	}

	return &GenerateDocumentResult{
		FileID: fileID,
	}, nil
}

type AddSignatureHandler struct {
	documents domain.DocumentRepository
}

func NewAddSignatureHandler(documents domain.DocumentRepository) *AddSignatureHandler {
	return &AddSignatureHandler{documents: documents}
}

type AddSignatureCommand struct {
	DocumentID string
	SignerName string
	SignerRole string
}

type AddSignatureResult struct {
	SignatureID string
}

func (h *AddSignatureHandler) Handle(ctx context.Context, cmd AddSignatureCommand) (*AddSignatureResult, error) {
	docID, err := domain.ParseDocumentID(cmd.DocumentID)
	if err != nil {
		return nil, fmt.Errorf("parse document ID: %w", err)
	}

	document, err := h.documents.FindByID(ctx, docID)
	if err != nil {
		return nil, fmt.Errorf("find document: %w", err)
	}

	signature, err := document.AddSignature(cmd.SignerName, cmd.SignerRole)
	if err != nil {
		return nil, fmt.Errorf("add signature: %w", err)
	}

	if err := h.documents.Save(ctx, document); err != nil {
		return nil, fmt.Errorf("save document: %w", err)
	}

	return &AddSignatureResult{
		SignatureID: signature.GetID().String(),
	}, nil
}

type SignDocumentHandler struct {
	documents domain.DocumentRepository
	bus       eventbus.Bus
}

func NewSignDocumentHandler(documents domain.DocumentRepository, bus eventbus.Bus) *SignDocumentHandler {
	return &SignDocumentHandler{
		documents: documents,
		bus:       bus,
	}
}

type SignDocumentCommand struct {
	DocumentID    string
	SignatureID   string
	SignatureData string
}

func (h *SignDocumentHandler) Handle(ctx context.Context, cmd SignDocumentCommand) error {
	docID, err := domain.ParseDocumentID(cmd.DocumentID)
	if err != nil {
		return fmt.Errorf("parse document ID: %w", err)
	}

	sigID, err := domain.ParseSignatureID(cmd.SignatureID)
	if err != nil {
		return fmt.Errorf("parse signature ID: %w", err)
	}

	document, err := h.documents.FindByID(ctx, docID)
	if err != nil {
		return fmt.Errorf("find document: %w", err)
	}

	if err := document.SignSignature(sigID, cmd.SignatureData); err != nil {
		return fmt.Errorf("sign document: %w", err)
	}

	if err := h.documents.Save(ctx, document); err != nil {
		return fmt.Errorf("save document: %w", err)
	}

	events := document.PullEvents()
	for _, event := range events {
		_ = h.bus.Publish(ctx, event)
	}

	return nil
}

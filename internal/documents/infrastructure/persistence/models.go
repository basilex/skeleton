package persistence

import (
	"time"

	"github.com/basilex/skeleton/internal/documents/domain"
)

type documentDTO struct {
	ID             string            `db:"id"`
	DocumentNumber string            `db:"document_number"`
	DocumentType   string            `db:"document_type"`
	ReferenceID    string            `db:"reference_id"`
	FileID         *string           `db:"file_id"`
	Status         string            `db:"status"`
	Metadata       map[string]string `db:"metadata"`
	CreatedAt      time.Time         `db:"created_at"`
	UpdatedAt      time.Time         `db:"updated_at"`
}

func (dto *documentDTO) toDomain(signatures []*signatureDTO) (*domain.Document, error) {
	id, err := domain.ParseDocumentID(dto.ID)
	if err != nil {
		return nil, err
	}

	var fileID *string
	if dto.FileID != nil {
		fileID = dto.FileID
	}

	status := domain.DocumentStatus(dto.Status)
	docType := domain.DocumentType(dto.DocumentType)

	domainSignatures := make([]*domain.Signature, 0, len(signatures))
	for _, sig := range signatures {
		domainSignatures = append(domainSignatures, sig.toDomain())
	}

	return domain.RestoreDocument(
		id,
		dto.DocumentNumber,
		docType,
		dto.ReferenceID,
		fileID,
		status,
		dto.Metadata,
		domainSignatures,
		[]domain.DocumentVersion{},
		domain.VersionNumber(1),
		dto.CreatedAt,
		dto.UpdatedAt,
	), nil
}

type signatureDTO struct {
	ID            string     `db:"id"`
	DocumentID    string     `db:"document_id"`
	SignerName    string     `db:"signer_name"`
	SignerRole    string     `db:"signer_role"`
	Status        string     `db:"status"`
	SignedAt      *time.Time `db:"signed_at"`
	SignatureData string     `db:"signature_data"`
	CreatedAt     time.Time  `db:"created_at"`
}

func (dto *signatureDTO) toDomain() *domain.Signature {
	id, _ := domain.ParseSignatureID(dto.ID)
	docID, _ := domain.ParseDocumentID(dto.DocumentID)
	status := domain.SignatureStatus(dto.Status)

	return domain.RestoreSignature(
		id,
		docID,
		dto.SignerName,
		dto.SignerRole,
		status,
		dto.SignedAt,
		dto.SignatureData,
		dto.CreatedAt,
	)
}

type templateDTO struct {
	ID           string    `db:"id"`
	Name         string    `db:"name"`
	DocumentType string    `db:"document_type"`
	Content      string    `db:"content"`
	Variables    []string  `db:"variables"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (dto *templateDTO) toDomain() *domain.Template {
	id, _ := domain.ParseTemplateID(dto.ID)
	docType := domain.DocumentType(dto.DocumentType)

	return domain.RestoreTemplate(
		id,
		dto.Name,
		docType,
		dto.Content,
		dto.Variables,
		dto.CreatedAt,
		dto.UpdatedAt,
	)
}

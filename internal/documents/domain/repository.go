package domain

import (
	"context"

	"github.com/basilex/skeleton/pkg/pagination"
)

type DocumentFilter struct {
	DocumentType  *DocumentType
	Status        *DocumentStatus
	ReferenceID   *string
	CreatedAfter  *string
	CreatedBefore *string
	Cursor        string
	Limit         int
}

type DocumentRepository interface {
	Save(ctx context.Context, document *Document) error
	FindByID(ctx context.Context, id DocumentID) (*Document, error)
	FindByDocumentNumber(ctx context.Context, documentNumber string) (*Document, error)
	FindByReferenceID(ctx context.Context, referenceID string) (*Document, error)
	FindAll(ctx context.Context, filter DocumentFilter) (pagination.PageResult[*Document], error)
	Delete(ctx context.Context, id DocumentID) error
}

type TemplateRepository interface {
	Save(ctx context.Context, template *Template) error
	FindByID(ctx context.Context, id TemplateID) (*Template, error)
	FindByDocumentType(ctx context.Context, documentType DocumentType) ([]*Template, error)
	FindAll(ctx context.Context) ([]*Template, error)
	Delete(ctx context.Context, id TemplateID) error
}

type SignatureRepository interface {
	Save(ctx context.Context, signature *Signature) error
	FindByID(ctx context.Context, id SignatureID) (*Signature, error)
	FindByDocumentID(ctx context.Context, documentID DocumentID) ([]*Signature, error)
	Delete(ctx context.Context, id SignatureID) error
}

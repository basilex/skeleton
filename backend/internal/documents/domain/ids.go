package domain

import (
	"github.com/basilex/skeleton/pkg/uuid"
)

type DocumentID uuid.UUID

func NewDocumentID() DocumentID {
	return DocumentID(uuid.NewV7())
}

func ParseDocumentID(s string) (DocumentID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return DocumentID{}, err
	}
	return DocumentID(id), nil
}

func (id DocumentID) String() string {
	return uuid.UUID(id).String()
}

func (id DocumentID) IsZero() bool {
	return uuid.UUID(id) == uuid.UUID{}
}

type TemplateID uuid.UUID

func NewTemplateID() TemplateID {
	return TemplateID(uuid.NewV7())
}

func ParseTemplateID(s string) (TemplateID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return TemplateID{}, err
	}
	return TemplateID(id), nil
}

func (id TemplateID) String() string {
	return uuid.UUID(id).String()
}

type SignatureID uuid.UUID

func NewSignatureID() SignatureID {
	return SignatureID(uuid.NewV7())
}

func ParseSignatureID(s string) (SignatureID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return SignatureID{}, err
	}
	return SignatureID(id), nil
}

func (id SignatureID) String() string {
	return uuid.UUID(id).String()
}

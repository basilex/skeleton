package domain

import "time"

type DocumentCreated struct {
	DocumentID     DocumentID
	DocumentType   DocumentType
	DocumentNumber string
	ReferenceID    string
	occurredAt     time.Time
}

func (e DocumentCreated) EventName() string {
	return "documents.document_created"
}

func (e DocumentCreated) OccurredAt() time.Time {
	return e.occurredAt
}

type DocumentGenerated struct {
	DocumentID     DocumentID
	DocumentNumber string
	FileID         string
	occurredAt     time.Time
}

func (e DocumentGenerated) EventName() string {
	return "documents.document_generated"
}

func (e DocumentGenerated) OccurredAt() time.Time {
	return e.occurredAt
}

type DocumentSigned struct {
	DocumentID     DocumentID
	DocumentNumber string
	SignerName     string
	occurredAt     time.Time
}

func (e DocumentSigned) EventName() string {
	return "documents.document_signed"
}

func (e DocumentSigned) OccurredAt() time.Time {
	return e.occurredAt
}

type SignatureSigned struct {
	DocumentID  DocumentID
	SignatureID SignatureID
	SignerName  string
	occurredAt  time.Time
}

func (e SignatureSigned) EventName() string {
	return "documents.signature_signed"
}

func (e SignatureSigned) OccurredAt() time.Time {
	return e.occurredAt
}

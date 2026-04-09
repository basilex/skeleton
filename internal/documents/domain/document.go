package domain

import (
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/eventbus"
)

type Document struct {
	id             DocumentID
	documentNumber string
	documentType   DocumentType
	referenceID    string
	fileID         *string
	status         DocumentStatus
	metadata       map[string]string
	signatures     []*Signature
	createdAt      time.Time
	updatedAt      time.Time
	events         []eventbus.Event
}

func NewDocument(
	documentNumber string,
	documentType DocumentType,
	referenceID string,
) (*Document, error) {
	if documentNumber == "" {
		return nil, fmt.Errorf("document number cannot be empty")
	}
	if !documentType.IsValid() {
		return nil, ErrInvalidDocumentType
	}

	now := time.Now()
	return &Document{
		id:             NewDocumentID(),
		documentNumber: documentNumber,
		documentType:   documentType,
		referenceID:    referenceID,
		status:         DocumentStatusDraft,
		metadata:       make(map[string]string),
		signatures:     make([]*Signature, 0),
		createdAt:      now,
		updatedAt:      now,
		events:         make([]eventbus.Event, 0),
	}, nil
}

func RestoreDocument(
	id DocumentID,
	documentNumber string,
	documentType DocumentType,
	referenceID string,
	fileID *string,
	status DocumentStatus,
	metadata map[string]string,
	signatures []*Signature,
	createdAt time.Time,
	updatedAt time.Time,
) *Document {
	return &Document{
		id:             id,
		documentNumber: documentNumber,
		documentType:   documentType,
		referenceID:    referenceID,
		fileID:         fileID,
		status:         status,
		metadata:       metadata,
		signatures:     signatures,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		events:         make([]eventbus.Event, 0),
	}
}

func (d *Document) GetID() DocumentID {
	return d.id
}

func (d *Document) GetDocumentNumber() string {
	return d.documentNumber
}

func (d *Document) GetDocumentType() DocumentType {
	return d.documentType
}

func (d *Document) GetReferenceID() string {
	return d.referenceID
}

func (d *Document) GetFileID() *string {
	return d.fileID
}

func (d *Document) GetStatus() DocumentStatus {
	return d.status
}

func (d *Document) GetMetadata() map[string]string {
	return d.metadata
}

func (d *Document) GetSignatures() []*Signature {
	return d.signatures
}

func (d *Document) GetCreatedAt() time.Time {
	return d.createdAt
}

func (d *Document) GetUpdatedAt() time.Time {
	return d.updatedAt
}

func (d *Document) SetFile(fileID string) {
	d.fileID = &fileID
	d.updatedAt = time.Now()
}

func (d *Document) MarkGenerated(fileID string) error {
	if d.status != DocumentStatusDraft {
		return fmt.Errorf("%w: can only generate draft documents", ErrInvalidDocumentStatus)
	}

	d.fileID = &fileID
	d.status = DocumentStatusGenerated
	d.updatedAt = time.Now()
	d.events = append(d.events, DocumentGenerated{
		DocumentID:     d.id,
		DocumentNumber: d.documentNumber,
		FileID:         fileID,
		occurredAt:     time.Now(),
	})
	return nil
}

func (d *Document) MarkSent() error {
	if d.status != DocumentStatusGenerated {
		return fmt.Errorf("%w: can only send generated documents", ErrInvalidDocumentStatus)
	}

	d.status = DocumentStatusSent
	d.updatedAt = time.Time{}
	return nil
}

func (d *Document) AddSignature(signerName string, signerRole string) (*Signature, error) {
	if d.status == DocumentStatusArchived {
		return nil, fmt.Errorf("%w: cannot add signature to archived document", ErrInvalidDocumentStatus)
	}

	signature, err := NewSignature(d.id, signerName, signerRole)
	if err != nil {
		return nil, err
	}

	d.signatures = append(d.signatures, signature)
	d.updatedAt = time.Now()
	return signature, nil
}

func (d *Document) SignSignature(signatureID SignatureID, signatureData string) error {
	for _, sig := range d.signatures {
		if sig.GetID() == signatureID {
			if err := sig.Sign(signatureData); err != nil {
				return err
			}
			d.updatedAt = time.Now()

			// Check if all signatures are signed
			allSigned := true
			for _, s := range d.signatures {
				if s.GetStatus() != SignatureStatusSigned {
					allSigned = false
					break
				}
			}

			if allSigned && d.status != DocumentStatusSigned {
				d.status = DocumentStatusSigned
				d.events = append(d.events, DocumentSigned{
					DocumentID:     d.id,
					DocumentNumber: d.documentNumber,
					SignerName:     sig.GetSignerName(),
					occurredAt:     time.Now(),
				})
			}

			return nil
		}
	}

	return ErrSignatureNotFound
}

func (d *Document) Archive() error {
	if d.status == DocumentStatusArchived {
		return fmt.Errorf("%w: document already archived", ErrInvalidDocumentStatus)
	}

	d.status = DocumentStatusArchived
	d.updatedAt = time.Now()
	return nil
}

func (d *Document) SetMetadata(key string, value string) {
	if d.metadata == nil {
		d.metadata = make(map[string]string)
	}
	d.metadata[key] = value
	d.updatedAt = time.Now()
}

func (d *Document) PullEvents() []eventbus.Event {
	events := d.events
	d.events = make([]eventbus.Event, 0)
	return events
}

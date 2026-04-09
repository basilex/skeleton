package domain

type DocumentType string

const (
	DocumentTypeInvoice     DocumentType = "invoice"
	DocumentTypeContract    DocumentType = "contract"
	DocumentTypeQuote       DocumentType = "quote"
	DocumentTypeOrder       DocumentType = "order"
	DocumentTypeReceipt     DocumentType = "receipt"
	DocumentTypeCertificate DocumentType = "certificate"
)

func (t DocumentType) String() string {
	return string(t)
}

func (t DocumentType) IsValid() bool {
	switch t {
	case DocumentTypeInvoice, DocumentTypeContract, DocumentTypeQuote,
		DocumentTypeOrder, DocumentTypeReceipt, DocumentTypeCertificate:
		return true
	default:
		return false
	}
}

type DocumentStatus string

const (
	DocumentStatusDraft     DocumentStatus = "draft"
	DocumentStatusGenerated DocumentStatus = "generated"
	DocumentStatusSent      DocumentStatus = "sent"
	DocumentStatusSigned    DocumentStatus = "signed"
	DocumentStatusArchived  DocumentStatus = "archived"
)

func (s DocumentStatus) String() string {
	return string(s)
}

func (s DocumentStatus) CanTransitionTo(newStatus DocumentStatus) bool {
	transitions := map[DocumentStatus][]DocumentStatus{
		DocumentStatusDraft:     {DocumentStatusGenerated, DocumentStatusArchived},
		DocumentStatusGenerated: {DocumentStatusSent, DocumentStatusSigned, DocumentStatusArchived},
		DocumentStatusSent:      {DocumentStatusSigned, DocumentStatusArchived},
		DocumentStatusSigned:    {DocumentStatusArchived},
		DocumentStatusArchived:  {},
	}

	allowed, exists := transitions[s]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}

	return false
}

type SignatureStatus string

const (
	SignatureStatusPending  SignatureStatus = "pending"
	SignatureStatusSigned   SignatureStatus = "signed"
	SignatureStatusRejected SignatureStatus = "rejected"
)

func (s SignatureStatus) String() string {
	return string(s)
}

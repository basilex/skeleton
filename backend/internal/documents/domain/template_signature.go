package domain

import (
	"time"
)

type Template struct {
	id           TemplateID
	name         string
	documentType DocumentType
	content      string
	variables    []string
	createdAt    time.Time
	updatedAt    time.Time
}

func NewTemplate(
	name string,
	documentType DocumentType,
	content string,
) (*Template, error) {
	if name == "" {
		return nil, ErrEmptyTemplateName
	}
	if content == "" {
		return nil, ErrEmptyTemplateContent
	}
	if !documentType.IsValid() {
		return nil, ErrInvalidDocumentType
	}

	now := time.Now()
	return &Template{
		id:           NewTemplateID(),
		name:         name,
		documentType: documentType,
		content:      content,
		variables:    extractVariables(content),
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func RestoreTemplate(
	id TemplateID,
	name string,
	documentType DocumentType,
	content string,
	variables []string,
	createdAt time.Time,
	updatedAt time.Time,
) *Template {
	return &Template{
		id:           id,
		name:         name,
		documentType: documentType,
		content:      content,
		variables:    variables,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (t *Template) GetID() TemplateID {
	return t.id
}

func (t *Template) GetName() string {
	return t.name
}

func (t *Template) GetDocumentType() DocumentType {
	return t.documentType
}

func (t *Template) GetContent() string {
	return t.content
}

func (t *Template) GetVariables() []string {
	return t.variables
}

func (t *Template) GetCreatedAt() time.Time {
	return t.createdAt
}

func (t *Template) GetUpdatedAt() time.Time {
	return t.updatedAt
}

func (t *Template) UpdateContent(content string) error {
	if content == "" {
		return ErrEmptyTemplateContent
	}
	t.content = content
	t.variables = extractVariables(content)
	t.updatedAt = time.Now()
	return nil
}

func extractVariables(content string) []string {
	// Simple extraction: find {{variable}} patterns
	// In production, use proper template parsing
	var vars []string
	// TODO: Implement variable extraction from template content
	return vars
}

type Signature struct {
	id            SignatureID
	documentID    DocumentID
	signerName    string
	signerRole    string
	status        SignatureStatus
	signedAt      *time.Time
	signatureData string
	createdAt     time.Time
}

func NewSignature(
	documentID DocumentID,
	signerName string,
	signerRole string,
) (*Signature, error) {
	if signerName == "" {
		return nil, ErrEmptySignerName
	}

	return &Signature{
		id:         NewSignatureID(),
		documentID: documentID,
		signerName: signerName,
		signerRole: signerRole,
		status:     SignatureStatusPending,
		createdAt:  time.Now(),
	}, nil
}

func RestoreSignature(
	id SignatureID,
	documentID DocumentID,
	signerName string,
	signerRole string,
	status SignatureStatus,
	signedAt *time.Time,
	signatureData string,
	createdAt time.Time,
) *Signature {
	return &Signature{
		id:            id,
		documentID:    documentID,
		signerName:    signerName,
		signerRole:    signerRole,
		status:        status,
		signedAt:      signedAt,
		signatureData: signatureData,
		createdAt:     createdAt,
	}
}

func (s *Signature) GetID() SignatureID {
	return s.id
}

func (s *Signature) GetDocumentID() DocumentID {
	return s.documentID
}

func (s *Signature) GetSignerName() string {
	return s.signerName
}

func (s *Signature) GetSignerRole() string {
	return s.signerRole
}

func (s *Signature) GetStatus() SignatureStatus {
	return s.status
}

func (s *Signature) GetSignedAt() *time.Time {
	return s.signedAt
}

func (s *Signature) GetSignatureData() string {
	return s.signatureData
}

func (s *Signature) GetCreatedAt() time.Time {
	return s.createdAt
}

func (s *Signature) Sign(signatureData string) error {
	if s.status == SignatureStatusSigned {
		return ErrSignatureAlreadySigned
	}
	if s.status == SignatureStatusRejected {
		return ErrSignatureRejected
	}

	now := time.Now()
	s.status = SignatureStatusSigned
	s.signedAt = &now
	s.signatureData = signatureData
	return nil
}

func (s *Signature) Reject(reason string) error {
	if s.status == SignatureStatusSigned {
		return ErrSignatureAlreadySigned
	}
	s.status = SignatureStatusRejected
	return nil
}

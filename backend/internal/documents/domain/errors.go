package domain

import (
	"errors"
)

var (
	ErrDocumentNotFound       = errors.New("document not found")
	ErrTemplateNotFound       = errors.New("template not found")
	ErrSignatureNotFound      = errors.New("signature not found")
	ErrInvalidDocumentType    = errors.New("invalid document type")
	ErrInvalidDocumentStatus  = errors.New("invalid document status")
	ErrDocumentAlreadySigned  = errors.New("document already signed")
	ErrDocumentNotGenerated   = errors.New("document not generated")
	ErrSignatureAlreadySigned = errors.New("signature already signed")
	ErrSignatureRejected      = errors.New("signature rejected")
	ErrEmptyTemplateName      = errors.New("template name cannot be empty")
	ErrEmptyTemplateContent   = errors.New("template content cannot be empty")
	ErrEmptySignerName        = errors.New("signer name cannot be empty")
)

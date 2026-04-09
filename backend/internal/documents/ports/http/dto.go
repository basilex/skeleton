package http

type CreateDocumentRequest struct {
	DocumentNumber string `json:"document_number" binding:"required"`
	DocumentType   string `json:"document_type" binding:"required"`
	ReferenceID    string `json:"reference_id"`
}

type GenerateDocumentRequest struct {
	TemplateID string            `json:"template_id" binding:"required"`
	Data       map[string]string `json:"data"`
}

type AddSignatureRequest struct {
	SignerName string `json:"signer_name" binding:"required"`
	SignerRole string `json:"signer_role"`
}

type SignDocumentRequest struct {
	SignatureID   string `json:"signature_id" binding:"required"`
	SignatureData string `json:"signature_data" binding:"required"`
}

type CreateTemplateRequest struct {
	Name         string `json:"name" binding:"required"`
	DocumentType string `json:"document_type" binding:"required"`
	Content      string `json:"content" binding:"required"`
}

type DocumentResponse struct {
	ID             string              `json:"id"`
	DocumentNumber string              `json:"document_number"`
	DocumentType   string              `json:"document_type"`
	ReferenceID    string              `json:"reference_id"`
	FileID         *string             `json:"file_id"`
	Status         string              `json:"status"`
	Metadata       map[string]string   `json:"metadata"`
	Signatures     []SignatureResponse `json:"signatures"`
	CreatedAt      string              `json:"created_at"`
	UpdatedAt      string              `json:"updated_at"`
}

type SignatureResponse struct {
	ID            string  `json:"id"`
	SignerName    string  `json:"signer_name"`
	SignerRole    string  `json:"signer_role"`
	Status        string  `json:"status"`
	SignedAt      *string `json:"signed_at"`
	SignatureData string  `json:"signature_data,omitempty"`
}

type TemplateResponse struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	DocumentType string   `json:"document_type"`
	Content      string   `json:"content"`
	Variables    []string `json:"variables"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

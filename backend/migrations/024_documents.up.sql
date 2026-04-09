-- Migration: 024_documents.up.sql
-- Description: Creates documents management tables for PDF generation and signing
-- Created: 2025-01-08

-- Document type enum
CREATE TYPE document_type AS ENUM (
    'invoice',
    'contract',
    'quote',
    'order',
    'receipt',
    'certificate'
);

-- Document status enum
CREATE TYPE document_status AS ENUM (
    'draft',
    'generated',
    'sent',
    'signed',
    'archived'
);

-- Signature status enum
CREATE TYPE signature_status AS ENUM (
    'pending',
    'signed',
    'rejected'
);

-- Templates table (reusable document templates)
CREATE TABLE document_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name VARCHAR(255) NOT NULL,
    document_type document_type NOT NULL,
    content TEXT NOT NULL,
    variables TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Documents table (generated documents)
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    document_number VARCHAR(100) NOT NULL UNIQUE,
    document_type document_type NOT NULL,
    reference_id VARCHAR(255),
    file_id UUID REFERENCES files(id) ON DELETE SET NULL,
    status document_status NOT NULL DEFAULT 'draft',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Signatures table (digital signatures for documents)
CREATE TABLE document_signatures (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    signer_name VARCHAR(255) NOT NULL,
    signer_role VARCHAR(255),
    status signature_status NOT NULL DEFAULT 'pending',
    signed_at TIMESTAMP WITH TIME ZONE,
    signature_data TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_documents_type ON documents(document_type);
CREATE INDEX idx_documents_status ON documents(status);
CREATE INDEX idx_documents_reference ON documents(reference_id);
CREATE INDEX idx_documents_created_at ON documents(created_at);
CREATE INDEX idx_document_templates_type ON document_templates(document_type);
CREATE INDEX idx_signatures_document ON document_signatures(document_id);
CREATE INDEX idx_signatures_status ON document_signatures(status);

-- Comments for documentation
COMMENT ON TABLE document_templates IS 'Reusable templates for document generation';
COMMENT ON TABLE documents IS 'Generated documents with versioning and signature tracking';
COMMENT ON TABLE document_signatures IS 'Digital signatures for documents';

COMMENT ON COLUMN documents.reference_id IS 'Reference to related entity (invoice_id, contract_id, etc.)';
COMMENT ON COLUMN documents.file_id IS 'Link to generated PDF file';
COMMENT ON COLUMN documents.metadata IS 'Document metadata (JSONB for flexibility)';
COMMENT ON COLUMN document_signatures.signature_data IS 'Digital signature data (base64)';
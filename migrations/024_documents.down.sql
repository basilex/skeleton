-- Migration: 024_documents.down.sql
-- Description: Drops documents management tables
-- Created: 2025-01-08

-- Drop in reverse order of creation
DROP TABLE IF EXISTS document_signatures CASCADE;
DROP TABLE IF EXISTS documents CASCADE;
DROP TABLE IF EXISTS document_templates CASCADE;

-- Drop enums
DROP TYPE IF EXISTS signature_status CASCADE;
DROP TYPE IF EXISTS document_status CASCADE;
DROP TYPE IF EXISTS document_type CASCADE;
-- ============================================================================
-- Skeleton Project — Global Schema Cleanup
-- Migration: 000_init_schema.down.sql
-- Description: Removes global functions and extension installed by init_schema
-- ============================================================================

-- Remove utility functions
DROP FUNCTION IF EXISTS uuid_v7_to_timestamp(UUID);
DROP FUNCTION IF EXISTS uuid_generate_v7();

-- Remove extension (optional, safe to keep)
DROP EXTENSION IF EXISTS "uuid-ossp";

-- ============================================================================
-- END OF MIGRATION 000_init_schema.down.sql
-- ============================================================================
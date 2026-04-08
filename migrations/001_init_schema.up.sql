-- ============================================================================
-- Skeleton Project — Global Schema Initialization
-- Migration: 000_init_schema.up.sql
-- Description: Installs the PostgreSQL extension and utility functions that
--              the rest of the schema depends on. Must run before any table
--              migrations so all defaults and helpers are available.
--
-- Contents
-- ────────
--   1. Extension   : uuid-ossp (legacy compat / ad-hoc use)
--   2. Function    : uuid_generate_v7()       — time-ordered UUIDv7 generator
--   3. Function    : uuid_v7_to_timestamp()   — extract timestamp from UUIDv7
-- ============================================================================

-- ============================================================================
-- 1. Extension
-- ============================================================================
-- uuid-ossp provides uuid_generate_v4() and related helpers.
-- The application generates all PKs itself (UUIDv7 via Go), but the extension
-- is useful for ad-hoc SQL, seed scripts, and cross-check queries.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- 2. uuid_generate_v7()
-- ============================================================================
-- Generates a new UUIDv7 (RFC 9562) value directly in PostgreSQL.
--
-- Algorithm
-- ─────────
--   1. Capture the current wall-clock time as a 48-bit millisecond timestamp.
--   2. Use gen_random_uuid() (built-in since PG 13, CSPRNG-backed) as the
--      source of 128 random bits.
--   3. Overwrite bytes 0–5 with the big-endian millisecond timestamp.
--   4. Set the version nibble  (bits 48–51) to 0x7.
--   5. Set the variant bits    (bits 64–65) to 0b10.
--   6. Return the result cast to UUID.
--
-- Why UUIDv7 as the default for new tables?
-- ──────────────────────────────────────────
--   • INSERT order matches B-tree index order → no page splits, lower WAL
--   • CLUSTER / VACUUM FULL produces time-ordered pages automatically
--   • Time-range queries on the PK column become index-range scans
--   • The embedded timestamp makes a separate created_at column redundant
--     for ordering (we still keep created_at for readability and filtering)
--
-- Note on monotonicity
-- ────────────────────
-- This function does NOT implement the sub-millisecond monotonic counter that
-- the Go generator (pkg/uuid) uses, because PostgreSQL functions run inside
-- transactions where strict intra-millisecond ordering within a single session
-- is rarely required. The application layer is the authoritative source of
-- UUIDs for all primary-key columns; this function exists for:
--   • Ad-hoc INSERT statements in migrations and seed scripts
--   • Data-loading scripts that run entirely inside the database
--   • A cross-check / fallback when the application layer is unavailable
--
-- Requirements: PostgreSQL 13+  (gen_random_uuid() is built-in, no pgcrypto)
-- ============================================================================
CREATE OR REPLACE FUNCTION uuid_generate_v7()
RETURNS UUID
LANGUAGE plpgsql
VOLATILE
PARALLEL SAFE
AS $$
DECLARE
    v_ms    BIGINT;   -- Unix time in milliseconds
    v_bytes BYTEA;    -- 16-byte working buffer
    v_hex   TEXT;     -- hex-encoded buffer for UUID formatting
BEGIN
    -- ── Step 1: 48-bit millisecond timestamp ─────────────────────────────────
    -- clock_timestamp() returns the actual wall clock (not the transaction
    -- start time), so multiple calls within the same transaction still produce
    -- distinct, increasing timestamps.
    v_ms := (extract(epoch FROM clock_timestamp()) * 1000)::BIGINT;

    -- ── Step 2: 128 random bits via gen_random_uuid() ────────────────────────
    -- Strip hyphens and decode to raw bytes so we can overwrite specific bytes.
    v_hex   := replace(gen_random_uuid()::TEXT, '-', '');
    v_bytes := decode(v_hex, 'hex');

    -- ── Step 3: Overwrite bytes 0–5 with the big-endian timestamp ────────────
    v_bytes := set_byte(v_bytes, 0, ((v_ms >> 40) & 255)::INT);
    v_bytes := set_byte(v_bytes, 1, ((v_ms >> 32) & 255)::INT);
    v_bytes := set_byte(v_bytes, 2, ((v_ms >> 24) & 255)::INT);
    v_bytes := set_byte(v_bytes, 3, ((v_ms >> 16) & 255)::INT);
    v_bytes := set_byte(v_bytes, 4, ((v_ms >>  8) & 255)::INT);
    v_bytes := set_byte(v_bytes, 5, ( v_ms        & 255)::INT);

    -- ── Step 4: Set version = 7 in the upper nibble of byte 6 ────────────────
    -- Byte 6 layout:  [ver=0111][rand_a bits 11:8]
    -- Mask lower 4 bits to preserve rand_a, then OR with 0x70.
    v_bytes := set_byte(v_bytes, 6,
        (get_byte(v_bytes, 6) & x'0f'::INT) | x'70'::INT
    );

    -- ── Step 5: Set variant = 10xx in the upper 2 bits of byte 7 ────────────
    -- Byte 7 layout:  [var=10][rand_b bits 63:56]
    -- Mask lower 6 bits to preserve rand_b, then OR with 0x80.
    -- NOTE: Corrected byte index from 8 to 7 (bytes are 0-indexed)
    v_bytes := set_byte(v_bytes, 7,
        (get_byte(v_bytes, 7) & x'3f'::INT) | x'80'::INT
    );

    -- ── Step 6: Format as standard UUID string and return ─────────────────────
    -- encode() produces a 32-char lowercase hex string; insert hyphens at the
    -- standard positions: 8-4-4-4-12.
    v_hex := encode(v_bytes, 'hex');

    RETURN (
        substring(v_hex,  1, 8) || '-' ||
        substring(v_hex,  9, 4) || '-' ||
        substring(v_hex, 13, 4) || '-' ||
        substring(v_hex, 17, 4) || '-' ||
        substring(v_hex, 21, 12)
    )::UUID;
END;
$$;

COMMENT ON FUNCTION uuid_generate_v7() IS
    'Generates a UUIDv7 (RFC 9562): 48-bit ms timestamp | ver=7 | 12-bit rand_a | variant | 62-bit rand_b. '
    'Time-ordered — suitable as a B-tree primary key. Requires PostgreSQL 13+ (no extensions).';

-- ============================================================================
-- 3. uuid_v7_to_timestamp(uuid)
-- ============================================================================
-- Extracts the embedded millisecond-precision UTC timestamp from a UUIDv7.
--
-- Example
-- ───────
--   SELECT uuid_v7_to_timestamp('018f1a2b-3c4d-7e5f-8a9b-0c1d2e3f4a5b');
--   → 2024-05-15 10:23:45.123+00
--
-- Useful for
-- ──────────
--   • Time-range filtering on PK columns without a separate created_at index
--   • Auditing / debugging ("when was this record created?")
--   • Expression indexes on UUID columns when millisecond precision is enough
--
-- The function is declared IMMUTABLE and STRICT:
--   IMMUTABLE — output depends only on input; safe in expression indexes
--   STRICT    — returns NULL for NULL input without executing the body
-- ============================================================================
CREATE OR REPLACE FUNCTION uuid_v7_to_timestamp(p_uuid UUID)
RETURNS TIMESTAMPTZ
LANGUAGE sql
IMMUTABLE
STRICT
PARALLEL SAFE
AS $$
    SELECT to_timestamp(
        (
            -- Reconstruct the 48-bit big-endian millisecond value from bytes 0–5.
            -- get_byte() is 0-indexed.
            (get_byte(decode(replace(p_uuid::TEXT, '-', ''), 'hex'), 0)::BIGINT << 40) |
            (get_byte(decode(replace(p_uuid::TEXT, '-', ''), 'hex'), 1)::BIGINT << 32) |
            (get_byte(decode(replace(p_uuid::TEXT, '-', ''), 'hex'), 2)::BIGINT << 24) |
            (get_byte(decode(replace(p_uuid::TEXT, '-', ''), 'hex'), 3)::BIGINT << 16) |
            (get_byte(decode(replace(p_uuid::TEXT, '-', ''), 'hex'), 4)::BIGINT <<  8) |
             get_byte(decode(replace(p_uuid::TEXT, '-', ''), 'hex'), 5)::BIGINT
        ) / 1000.0   -- convert milliseconds → seconds (to_timestamp expects seconds)
    ) AT TIME ZONE 'UTC';
$$;

COMMENT ON FUNCTION uuid_v7_to_timestamp(UUID) IS
    'Extracts the 48-bit millisecond-precision UTC creation timestamp embedded in a UUIDv7. '
    'Returns NULL for NULL input (STRICT). IMMUTABLE — safe for use in expression indexes.';

-- ============================================================================
-- END OF MIGRATION 000_init_schema.up.sql
-- ============================================================================
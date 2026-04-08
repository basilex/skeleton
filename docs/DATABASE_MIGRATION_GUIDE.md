# Database Migration Guide

Complete guide for database migrations, PostgreSQL 16 features, and best practices.

## Table of Contents

1. [Overview](#overview)
2. [Migration Architecture](#migration-architecture)
3. [UUID v7 Implementation](#uuid-v7-implementation)
4. [Migration Workflow](#migration-workflow)
5. [PostgreSQL 16 Features](#postgresql-16-features)
6. [Best Practices](#best-practices)
7. [Troubleshooting](#troubleshooting)

## Overview

### Technology Stack

- **Database**: PostgreSQL 16 ONLY
- **Driver**: pgx/v5 with pgxpool (no ORM)
- **Migrations**: Custom Go tool (`scripts/migrate`)
- **IDs**: UUID v7 (time-sortable, RFC 9562)
- **JSONB**: For all metadata fields

### Why PostgreSQL 16?

| Feature | Benefit |
|---------|---------|
| **UUID v7 Support** | Time-sortable IDs, instant creation time extraction |
| **JSONB** | Queryable JSON with GIN indexes (10-100x faster) |
| **Generated Columns** | Computed values without triggers |
| **Materialized Views** | Fast aggregations |
| **Parallel Queries** | Better performance on large datasets |
| **pg_stat_statements** | Query performance monitoring |

### Why UUID v7?

| UUID Version | Sortable? | Storage | Time Extraction |
|--------------|-----------|----------|-----------------|
| UUID v4 | ❌ Random | 36 bytes (TEXT) | ❌ Impossible |
| UUID v7 | ✅ Time-ordered | 16 bytes (UUID) | ✅ Instant |

**Benefits of UUID v7:**
- **B-tree friendly**: Sequential inserts (no page splits)
- **Time-range queries**: `WHERE id BETWEEN '...' AND '...'`
- **Instant timestamp**: `uuid_v7_to_timestamp(id)` shows creation time
- **Distributed friendly**: No coordination needed
- **56% storage reduction**: 16 bytes vs 36 bytes

## Migration Architecture

### Migration Tool

Custom Go migration tool (`scripts/migrate`):

```go
// Simple migration runner
go run ./scripts/migrate -action=up      // Apply migrations
go run ./scripts/migrate -action=down    // Rollback last migration
go run ./scripts/migrate -action=status  // Check status
```

### Migration Files

Sequential numbering starting from 001:

```
migrations/
├── 001_init_schema.up.sql           # UUID v7 functions
├── 001_init_schema.down.sql
├── 002_users.up.sql                # Users table
├── 002_users.down.sql
├── 003_roles.up.sql                # Roles table
├── 003_roles.down.sql
...
├── 017_file_processings.up.sql      # Last table
└── 017_file_processings.down.sql
```

### Migration Table

```sql
CREATE TABLE schema_migrations (
    version    INTEGER PRIMARY KEY,
    dirty      BOOLEAN NOT NULL DEFAULT false,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Migration Rules

1. **Always use sequential numbering**: 001, 002, 003...
2. **Always provide up and down**: `xxx.up.sql` and `xxx.down.sql`
3. **Use UUID v7 for IDs**: `DEFAULT uuid_generate_v7()`
4. **Use JSONB for metadata**: `JSONB DEFAULT '{}'`
5. **Never use TEXT for IDs or timestamps**
6. **Always add GIN index for JSONB**

## UUID v7 Implementation

### PostgreSQL Function

Migration 001 creates the UUID v7 generator:

```sql
CREATE OR REPLACE FUNCTION uuid_generate_v7()
RETURNS UUID
LANGUAGE plpgsql
AS $$
DECLARE
    v_ms    BIGINT;
    v_bytes BYTEA;
    v_hex   TEXT;
BEGIN
    -- Get current timestamp in milliseconds
    v_ms := (EXTRACT(EPOCH FROM clock_timestamp()) * 1000)::BIGINT;
    
    -- Get random bytes from gen_random_uuid()
    v_hex := replace(gen_random_uuid()::TEXT, '-', '');
    v_bytes := decode(v_hex, 'hex');
    
    -- Overwrite bytes 0-5 with timestamp
    v_bytes := set_byte(v_bytes, 0, ((v_ms >> 40) & 255)::INT);
    v_bytes := set_byte(v_bytes, 1, ((v_ms >> 32) & 255)::INT);
    v_bytes := set_byte(v_bytes, 2, ((v_ms >> 24) & 255)::INT);
    v_bytes := set_byte(v_bytes, 3, ((v_ms >> 16) & 255)::INT);
    v_bytes := set_byte(v_bytes, 4, ((v_ms >>  8) & 255)::INT);
    v_bytes := set_byte(v_bytes, 5, (v_ms & 255)::INT);
    
    -- Set version = 7 in byte 6
    v_bytes := set_byte(v_bytes, 6, (get_byte(v_bytes, 6) & 15) | 112);
    
    -- Set variant = 10xx in byte 7
    v_bytes := set_byte(v_bytes, 7, (get_byte(v_bytes, 7) & 63) | 128);
    
    -- Return as UUID
    v_hex := encode(v_bytes, 'hex');
    RETURN (substring(v_hex, 1, 8) || '-' ||
            substring(v_hex, 9, 4) || '-' ||
            substring(v_hex, 13, 4) || '-' ||
            substring(v_hex, 17, 4) || '-' ||
            substring(v_hex, 21, 12))::UUID;
END;
$$;
```

### Timestamp Extraction

```sql
-- Extract creation time from UUID v7
CREATE OR REPLACE FUNCTION uuid_v7_to_timestamp(p_uuid UUID)
RETURNS TIMESTAMPTZ
LANGUAGE sql
IMMUTABLE
STRICT
AS $$
    SELECT to_timestamp(
        ((get_byte(decode(replace(p_uuid::TEXT, '-', ''), 'hex'), 0)::BIGINT << 40) |
         (get_byte(decode(replace(p_uuid::TEXT, '-', ''), 'hex'), 1)::BIGINT << 32) |
         (get_byte(decode(replace(p_uuid::TEXT, '-', ''), 'hex'), 2)::BIGINT << 24) |
         (get_byte(decode(replace(p_uuid::TEXT, '-', ''), 'hex'), 3)::BIGINT << 16) |
         (get_byte(decode(replace(p_uuid::TEXT, '-', ''), 'hex'), 4)::BIGINT <<  8) |
          get_byte(decode(replace(p_uuid::TEXT, '-', ''), 'hex'), 5)::BIGINT) / 1000.0
    ) AT TIME ZONE 'UTC';
$$;
```

### Usage Examples

```sql
-- Generate UUID v7
SELECT uuid_generate_v7();
-- Result: 019d6d0b-1234-7abc-8def-123456789abc
--         ^^^^^^^^ ^-- version 7
--         timestamp  ^-- variant 10xx

-- Extract timestamp
SELECT uuid_v7_to_timestamp('019d6d0b-1234-7abc-8def-123456789abc');
-- Result: 2026-04-08 12:34:56.789+00

-- Use as primary key
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Time-range queries
SELECT * FROM users
WHERE id >= '019d6d0b-0000-7000-8000-000000000000'
  AND id <= '019d6d0b-ffff-7fff-bfff-ffffffffffff';
```

### Go Integration

```go
// pkg/uuid/uuid.go
package uuid

import (
    "crypto/rand"
    "time"
)

type UUID [16]byte

func NewV7() UUID {
    var u UUID
    
    // Get current timestamp in milliseconds
    ms := uint64(now.UnixMilli())
    
    // Fill timestamp bytes 0-5
    u[0] = byte(ms >> 40)
    u[1] = byte(ms >> 32)
    u[2] = byte(ms >> 24)
    u[3] = byte(ms >> 16)
    u[4] = byte(ms >> 8)
    u[5] = byte(ms)
    
    // Set version 7 in byte 6
    u[6] = 0x70 | (byte(randUint16()) >> 4)
    
    // Set variant 10xx in byte 7
    u[7] = 0x80 | (get_byte(rand_bytes, 1) | 128)
    
    // Random bytes 8-15
    rand.Read(u[8:])
    
    return u
}

// Parse converts string to UUID
func Parse(s string) (UUID, error) {
    // Implementation...
}
```

## Migration Workflow

### Create New Migration

```bash
# Create files
touch migrations/018_new_table.up.sql
touch migrations/018_new_table.down.sql
```

### Write Up Migration

```sql
-- migrations/018_new_table.up.sql
CREATE TABLE new_table (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name TEXT NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- UUID foreign key
CREATE INDEX idx_new_table_user ON new_table(user_id);

-- GIN index for JSONB
CREATE INDEX idx_new_table_metadata_gin ON new_table USING GIN (metadata);

-- Partial index for active records
CREATE INDEX idx_new_table_active ON new_table(id) WHERE deleted_at IS NULL;

-- Generated column (bonus)
ALTER TABLE new_table ADD COLUMN name_lower TEXT
    GENERATED ALWAYS AS (LOWER(name)) STORED;
```

### Write Down Migration

```sql
-- migrations/018_new_table.down.sql
DROP TABLE IF EXISTS new_table CASCADE;
```

### Apply Migrations

```bash
# Apply single migration
make migrate-up

# Check status
make migrate-status
# Current version: 18 (dirty: false)

# Verify table
make db-tables
# List of tables...

# Check structure
make psql
\d new_table
```

### Rollback Migration

```bash
# Rollback last migration
make migrate-down

# Rollback multiple (manual)
for i in {1..5}; do make migrate-down; done

# Reset everything
make migrate-reset
```

## PostgreSQL 16 Features

### JSONB with GIN Indexes

```sql
-- Create table with JSONB
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    filename TEXT NOT NULL,
    metadata JSONB DEFAULT '{}'
);

-- GIN index for JSON queries
CREATE INDEX idx_files_metadata_gin ON files USING GIN (metadata);

-- Query examples
SELECT * FROM files WHERE metadata @> '{"width": 1920}';
SELECT * FROM files WHERE metadata->>'format' = 'PNG';
SELECT * FROM files WHERE metadata ? 'thumbnail';
SELECT * FROM files WHERE metadata @> '{"tags": ["landscape"]}';

-- Index-only scan
EXPLAIN ANALYZE
SELECT * FROM files WHERE metadata @> '{"width": 1920}';
```

### Generated Columns

```sql
-- Computed fields without triggers
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    price_cents INTEGER NOT NULL,
    currency TEXT NOT NULL DEFAULT 'USD',
    price_formatted TEXT
        GENERATED ALWAYS AS (
            CASE currency
                WHEN 'USD' THEN '$' || (price_cents / 100.0)::TEXT
                WHEN 'EUR' THEN '€' || (price_cents / 100.0)::TEXT
                ELSE (price_cents / 100.0)::TEXT || ' ' || currency
            END
        ) STORED,
    file_extension VARCHAR(10)
        GENERATED ALWAYS AS (LOWER(SPLIT_PART(filename, '.', -1))) STORED
);
```

### Materialized Views

```sql
-- Fast aggregations
CREATE MATERIALIZED VIEW user_stats AS
SELECT 
    user_id,
    COUNT(*) AS total_files,
    SUM(size) AS total_bytes,
    AVG(size) AS avg_file_size,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '7 days') AS files_last_week
FROM files
GROUP BY user_id;

CREATE UNIQUE INDEX idx_user_stats_user ON user_stats (user_id);

-- Refresh periodically
REFRESH MATERIALIZED VIEW CONCURRENTLY user_stats;

-- Query
SELECT * FROM user_stats WHERE user_id = 'uuid-here';
```

### Partial Indexes

```sql
-- Index only active records
CREATE INDEX idx_users_active ON users(email) WHERE is_active = TRUE;

-- Index only expired files
CREATE INDEX idx_files_expired ON files(id)
    WHERE expires_at IS NOT NULL AND expires_at < NOW();

-- Index pending tasks
CREATE INDEX idx_tasks_pending ON tasks(id)
    WHERE status = 'pending' AND scheduled_at < NOW();
```

### Composite Indexes

```sql
-- Multiple columns in one index
CREATE INDEX idx_tasks_user_status_created ON tasks(user_id, status, created_at);

-- Query optimization
SELECT * FROM tasks
WHERE user_id = 'uuid' AND status = 'pending'
ORDER BY created_at DESC;
-- Uses index scan
```

## Best Practices

### DO ✅

```sql
-- ✅ Use UUID v7 for primary keys
id UUID PRIMARY KEY DEFAULT uuid_generate_v7()

-- ✅ Use TIMESTAMPTZ for timestamps
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

-- ✅ Use JSONB for flexible metadata
metadata JSONB DEFAULT '{}'

-- ✅ Add GIN index for JSONB
CREATE INDEX idx_table_metadata_gin ON table USING GIN (metadata);

-- ✅ Use foreign keys with UUID
user_id UUID REFERENCES users(id) ON DELETE CASCADE

-- ✅ Add partial indexes for common queries
CREATE INDEX idx_users_active ON users(email) WHERE is_active = TRUE;

-- ✅ Use generated columns for computed values
file_extension VARCHAR(10) GENERATED ALWAYS AS (LOWER(SPLIT_PART(filename, '.', -1))) STORED

-- ✅ Create materialized views for aggregations
CREATE MATERIALIZED VIEW user_stats AS ...

-- ✅ Use check constraints for data integrity
CHECK (status IN ('pending', 'processing', 'completed', 'failed'))

-- ✅ Add comments for documentation
COMMENT ON TABLE users IS 'Application users with UUID v7 primary key';
COMMENT ON COLUMN users.metadata IS 'JSONB field - queryable with @> operator';
```

### DON'T ❌

```sql
-- ❌ Don't use TEXT for IDs
id TEXT PRIMARY KEY  -- Use UUID instead

-- ❌ Don't use TEXT for timestamps
created_at TEXT  -- Use TIMESTAMPTZ instead

-- ❌ Don't use TEXT for JSON
metadata TEXT  -- Use JSONB instead

-- ❌ Don't use UUID v4 (random)
id UUID DEFAULT gen_random_uuid()  -- Use uuid_generate_v7() instead

-- ❌ Don't forget GIN index for JSONB
CREATE TABLE files (..., metadata JSONB);  -- Missing GIN index!

-- ❌ Don't use INTEGER for boolean
is_active INTEGER  -- Use BOOLEAN instead

-- ❌ Don't create indexes without WHERE clause for partial data
CREATE INDEX idx_files_expired ON files(expires_at);  -- Use partial index
-- Better:
CREATE INDEX idx_files_expired ON files(id) WHERE expires_at IS NOT NULL;
```

### Naming Conventions

```sql
-- Tables: snake_case, plural
users, files, notifications

-- Columns: snake_case
created_at, user_id, is_active

-- Primary keys: simple `id`
id UUID PRIMARY KEY

-- Foreign keys: `{table}_id`
user_id UUID REFERENCES users(id)

-- Indexes: `idx_{table}_{columns}`
idx_users_email
idx_files_user_status
idx_tasks_scheduled

-- Partial indexes: `idx_{table}_{purpose}`
idx_users_active
idx_files_expired
idx_tasks_pending

-- GIN indexes: `idx_{table}_{column}_gin`
idx_files_metadata_gin
idx_tasks_payload_gin
```

## Migration Patterns

### Add Column

```sql
-- Up
ALTER TABLE users ADD COLUMN phone TEXT;

-- Down
ALTER TABLE users DROP COLUMN phone;
```

### Add Column with Default

```sql
-- Up
ALTER TABLE users ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT FALSE;

-- Down
ALTER TABLE users DROP COLUMN is_verified;
```

### Add Foreign Key

```sql
-- Up
ALTER TABLE files ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX idx_files_user ON files(user_id);

-- Down
DROP INDEX idx_files_user;
ALTER TABLE files DROP COLUMN user_id;
```

### Change Column Type

```sql
-- Up (TEXT to JSONB)
ALTER TABLE files ADD COLUMN metadata_new JSONB;
UPDATE files SET metadata_new = metadata::jsonb WHERE metadata IS NOT NULL;
ALTER TABLE files DROP COLUMN metadata;
ALTER TABLE files RENAME COLUMN metadata_new TO metadata;

-- Down (JSONB to TEXT)
ALTER TABLE files ADD COLUMN metadata_new TEXT;
UPDATE files SET metadata_new = metadata::text WHERE metadata IS NOT NULL;
ALTER TABLE files DROP COLUMN metadata;
ALTER TABLE files RENAME COLUMN metadata_new TO metadata;
```

### Add Index

```sql
-- Up
CREATE INDEX CONCURRENTLY idx_users_email_lower ON users(LOWER(email));

-- Down (if created CONCURRENTLY, drop normally)
DROP INDEX idx_users_email_lower;
```

### Add Generated Column

```sql
-- Up
ALTER TABLE files ADD COLUMN file_extension VARCHAR(10)
    GENERATED ALWAYS AS (LOWER(SPLIT_PART(filename, '.', -1))) STORED;

-- Down
ALTER TABLE files DROP COLUMN file_extension;
```

## Troubleshooting

### Migration Failed

```bash
# Check error message
make migrate-up
# ERROR: relation "users" already exists

# Check current status
make migrate-status
# Current version: 5 (dirty: true)

# Manual fix
make psql
DELETE FROM schema_migrations WHERE version = 5;
\q

# Retry migration
make migrate-up
```

### UUID v7 Not Working

```sql
-- Check if function exists
\df uuid_generate_v7

-- Test generation
SELECT uuid_generate_v7();

-- If missing, re-run init migration
make migrate-down  -- rollback all
make migrate-up    -- re-apply all
```

### JSONB Query Slow

```sql
-- Check if GIN index exists
SELECT indexname FROM pg_indexes 
WHERE tablename = 'files' AND indexname LIKE '%gin%';

-- If missing, create it
CREATE INDEX idx_files_metadata_gin ON files USING GIN (metadata);

-- Check query plan
EXPLAIN ANALYZE
SELECT * FROM files WHERE metadata @> '{"width": 1920}';

-- Should show: "Index Scan using idx_files_metadata_gin"
```

### Foreign Key Violation

```sql
-- Check existing data
SELECT * FROM files WHERE user_id NOT IN (SELECT id FROM users);

-- Fix invalid foreign keys
-- Option 1: Delete orphaned records
DELETE FROM files WHERE user_id NOT IN (SELECT id FROM users);

-- Option 2: Set to NULL (if allowed)
UPDATE files SET user_id = NULL WHERE user_id NOT IN (SELECT id FROM users);

-- Option 3: Update to valid value
UPDATE files SET user_id = 'uuid-here' WHERE user_id NOT IN (SELECT id FROM users);
```

### Migration Lock Timeout

```bash
# Check for long-running transactions
make db-connections

# Kill blocking transaction
SELECT pg_cancel_backend(pid) FROM pg_stat_activity 
WHERE datname = 'skeleton' AND state = 'active';

# Retry migration
make migrate-up
```

### Performance Issues

```sql
-- Check slow queries
SELECT query, calls, mean_time 
FROM pg_stat_statements 
ORDER BY mean_time DESC LIMIT 20;

-- Check missing indexes
SELECT schemaname, tablename, attname, n_distinct, correlation
FROM pg_stats
WHERE n_distinct > 100 AND correlation < 0.1;

-- Check table bloat
SELECT schemaname, tablename, 
       pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Vacuum analyze
VACUUM ANALYZE;
```

## Data Migration Scripts

### Bulk UUID Conversion

```sql
-- If you have TEXT UUIDs, convert to UUID type
ALTER TABLE users ADD COLUMN id_new UUID DEFAULT uuid_generate_v7();
UPDATE users SET id_new = id::uuid;
ALTER TABLE users DROP COLUMN id;
ALTER TABLE users RENAME COLUMN id_new TO id;
```

### Bulk JSONB Conversion

```sql
-- If you have TEXT JSON, convert to JSONB
ALTER TABLE files ADD COLUMN metadata_new JSONB;
UPDATE files SET metadata_new = metadata::jsonb WHERE metadata IS NOT NULL;
UPDATE files SET metadata_new = '{}' WHERE metadata_new IS NULL;
ALTER TABLE files ALTER COLUMN metadata_new SET NOT NULL;
ALTER TABLE files DROP COLUMN metadata;
ALTER TABLE files RENAME COLUMN metadata_new TO metadata;
```

---

**Related Documentation:**
- [DEVELOPMENT.md](DEVELOPMENT.md) - Development workflow
- [Makefile Guide](../MAKEFILE_REFERENCE.md) - All Makefile commands
- [ADR-016 Database Stack](adr/ADR-016-database-stack.md) - Database architecture decisions
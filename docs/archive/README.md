# Archive: PostgreSQL Migration Documents

This directory contains historical documents from the SQLite to PostgreSQL migration.

## What's Archived Here

- **CODE_REVIEW_POSTGRES_MIGRATION.md** - Code review findings during migration
- **DOCKER_CHANGES.md** - Docker environment changes during migration
- **POSTGRES_MIGRATION_SUMMARY.md** - Complete summary of migration work
- **SCHEMA_MIGRATION_PLAN.md** - Schema migration planning document

## Why These Were Archived

These documents served their purpose during the migration phase. The project now runs exclusively on PostgreSQL 16. For current documentation, see:

- **README.md** - Project overview and getting started
- **EXECUTION_PLAN.md** - Deployment and operational procedures
- **docs/DEPLOYMENT_GUIDE.md** - Comprehensive deployment guide
- **docs/adr/ADR-016-pgx-with-postgres.md** - Architecture decision for PostgreSQL

## Current State

- **Database**: PostgreSQL 16 ONLY (no SQLite support)
- **Driver**: Pure pgx/pqxpool (no sqlx, no GORM)
- **IDs**: Native UUID v7 (16 bytes vs 36 bytes TEXT)
- **Metadata**: Native JSONB (queryable with GIN indexes)
- **Status**: Migration complete, all tests passing

## Migration Achievements

1. ✅ All repositories migrated to pgxpool
2. ✅ Domain types use UUID instead of string
3. ✅ JSONB columns with GIN indexes
4. ✅ Generated columns for computed fields
5. ✅ Materialized views for aggregations
6. ✅ PostgreSQL functions for complex operations
7. ✅ Comprehensive indexes (40+ indexes)
8. ✅ Testcontainers for integration tests
9. ✅ Benchmark suite for performance
10. ✅ Monitoring queries for production

## For Reference

If you need to understand the migration journey or decisions made, these archived documents provide detailed context. For day-to-day operations, use the current documentation mentioned above.

---

Migration completed: April 2026  
Current version: PostgreSQL 16 with pure pgx  
Architecture: See docs/adr/ for full ADR history
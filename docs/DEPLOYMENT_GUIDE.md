# PostgreSQL Deployment Guide

Complete guide for deploying and maintaining PostgreSQL in all environments.

## Overview

This guide covers:
1. **Staging Deployment** - Test migrations before production
2. **Benchmark Execution** - Gather performance baselines
3. **Production Monitoring** - Monitor database health
4. **Index Optimization** - Ongoing performance tuning

## Prerequisites

```bash
# Required tools
- PostgreSQL 16+ client (psql)
- Go 1.24+
- Docker (for local testing)

# Environment variables
export DATABASE_URL="postgres://user:pass@host:5432/db"
export STAGING_DATABASE_URL="postgres://user:pass@staging:5432/db"
export BENCHMARK_DATABASE_URL="postgres://user:pass@localhost:5432/benchmark"
```

## 1. Staging Deployment

### Quick Start

```bash
# Deploy to staging with full checks
./scripts/deploy-staging.sh
```

### What It Does

1. ✅ Pre-deployment checks
   - Database connectivity
   - Migration status
   - Pending migrations

2. ✅ Backup creation
   - Schema backup
   - Critical data backup
   - Transaction log

3. ✅ Migration execution
   - Run in transactions
   - Automatic rollback on failure
   - Version tracking

4. ✅ Schema verification
   - Index count check
   - Extension verification
   - Materialized view validation

5. ✅ Performance baseline
   - Query timing tests
   - Connection pool check

6. ✅ Health checks
   - Table sizes
   - Index usage
   - Cache ratios

7. ✅ Post-deployment
   - Materialized view refresh
   - Table analysis
   - Report generation

### Deployment Checklist

- [ ] Review migration files
- [ ] Test in local environment
- [ ] Create database backup
- [ ] Run deployment script
- [ ] Verify schema changes
- [ ] Run application tests
- [ ] Monitor for 24 hours
- [ ] Document any issues

### Rollback Procedure

```bash
# If deployment fails, rollback:
psql $STAGING_DATABASE_URL < backups/staging/YYYYMMDD_HHMMSS/schema.sql

# Restore data
for table in users roles files tasks notifications; do
    psql $STAGING_DATABASE_URL < backups/staging/YYYYMMDD_HHMMSS/${table}_data.sql
done

# Restart application
docker-compose restart
```

## 2. Performance Benchmarks

### Run Benchmarks

```bash
# Run all benchmarks
./scripts/run-benchmarks.sh

# Run specific benchmark
./scripts/run-benchmarks.sh BenchmarkGetFileByID

# With custom database
BENCHMARK_DATABASE_URL="postgres://..." ./scripts/run-benchmarks.sh
```

### Benchmark Categories

| Benchmark | Target | Description |
|-----------|--------|-------------|
| BenchmarkGetFileByID | < 1ms | Primary key lookup |
| BenchmarkListFilesByOwner | < 5ms | Index scan with limit |
| BenchmarkSearchFilesWithJSONB | < 10ms | GIN index query |
| BenchmarkTaskQueue | < 2ms | FOR UPDATE SKIP LOCKED |
| BenchmarkNotificationBatch | < 50ms | Batch insert |
| BenchmarkAuditLogInsert | < 1ms | Simple insert |

### Benchmark Workflow

1. **Setup**: Creates test database with seed data
   - 1,000 users
   - 10,000 files
   - 50,000 tasks
   - 20,000 notifications

2. **Execute**: Runs each benchmark 5 times
   - Measures latency
   - Tracks memory allocation
   - Records throughput

3. **Compare**: Compares with baseline
   - Detects regressions (>20% slower)
   - Generates comparison report

4. **Report**: Creates detailed report
   - Raw benchmarks
   - Database statistics
   - Performance targets
   - Recommendations

### Continuous Benchmarking

Add to CI/CD pipeline:

```yaml
# .github/workflows/benchmark.yml
name: Benchmark

on:
  schedule:
    - cron: '0 2 * * *'  # Nightly at 2 AM
  push:
    branches: [main]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_DB: skeleton_benchmark
    steps:
      - uses: actions/checkout@v3
      - run: ./scripts/run-benchmarks.sh
      - uses: actions/upload-artifact@v3
        with:
          name: benchmark-results
          path: benchmark_results/
```

## 3. Production Monitoring

### Monitoring Dashboard Queries

Located in `docs/monitoring/postgres_monitoring_queries.sql`

#### Key Metrics to Monitor

**1. Connection Health**
```sql
-- Active connections by state
SELECT datname, state, COUNT(*) 
FROM pg_stat_activity 
GROUP BY datname, state;
```

**2. Index Performance**
```sql
-- Index hit ratio (should be > 99%)
SELECT round((sum(idx_blks_hit)::float / 
       (sum(idx_blks_hit) + sum(idx_blks_read))) * 100, 2)
FROM pg_statio_user_indexes;
```

**3. Query Performance**
```sql
-- Slow queries (> 100ms)
SELECT query, calls, mean_time
FROM pg_stat_statements
WHERE mean_time > 100
ORDER BY mean_time DESC;
```

**4. Table Bloat**
```sql
-- Dead tuple ratio (should be < 20%)
SELECT n_dead_tup::float / n_live_tup * 100
FROM pg_stat_user_tables
WHERE n_live_tup > 1000;
```

**5. Cache Hit Ratio**
```sql
-- Should be > 99%
SELECT sum(heap_blks_hit) / 
       (sum(heap_blks_hit) + sum(heap_blks_read))
FROM pg_statio_user_tables;
```

### Grafana Dashboard

Import queries into Grafana for visualization:

```json
{
  "dashboard": "PostgreSQL Performance",
  "panels": [
    {"title": "Connection Count", "query": "SELECT COUNT(*) FROM pg_stat_activity"},
    {"title": "Index Hit Ratio", "query": "SELECT ... FROM pg_statio_user_indexes"},
    {"title": "Query Latency", "query": "SELECT ... FROM pg_stat_statements"},
    {"title": "Table Bloat", "query": "SELECT ... FROM pg_stat_user_tables"}
  ]
}
```

### Alerting Rules

Configure alerts for:

| Metric | Threshold | Severity |
|--------|-----------|----------|
| Connection Count | > 80% max | Warning |
| Index Hit Ratio | < 95% | Critical |
| Slow Queries | > 1s | Warning |
| Table Bloat | > 20% | Warning |
| Cache Hit Ratio | < 95% | Critical |
| Replication Lag | > 10s | Critical |

### Monitoring Checklist

- [ ] Setup pg_stat_statements extension
- [ ] Configure connection pooler (pgBouncer)
- [ ] Enable slow query log
- [ ] Setup Grafana dashboards
- [ ] Configure alerting rules
- [ ] Schedule daily vacuum/analyze
- [ ] Monitor disk space usage

## 4. Index Optimization

### Run Analysis

```bash
# Analyze indexes (dry-run)
./scripts/optimize-indexes.sh --dry-run

# Apply optimizations
./scripts/optimize-indexes.sh
```

### What It Analyzes

1. **Unused Indexes**
   - Zero scans since last reset
   - Safe to drop (except primary keys)
   - Saves disk space and write performance

2. **Duplicate Indexes**
   - Same columns, same type
   - Wasted disk space
   - Unnecessary write overhead

3. **Missing Indexes**
   - High sequential scans
   - Low index scan ratio
   - Candidate for new index

4. **Index Bloat**
   - Large indexes with many updates
   - Decreased performance
   - Candidate for REINDEX

5. **JSONB Indexes**
   - GIN indexes for metadata fields
   - Critical for JSON query performance

### Optimization Workflow

```bash
# Weekly scheduled job
0 3 * * 0 /path/to/scripts/optimize-indexes.sh > /var/log/postgres-index-analysis.log 2>&1
```

### Index Recommendations

**Drop Unused Indexes:**
```sql
-- Check if really unused
SELECT * FROM pg_stat_user_indexes WHERE indexname = 'idx_name';

-- Drop carefully
DROP INDEX CONCURRENTLY schema.idx_name;
```

**Add Missing Indexes:**
```sql
-- Analyze query pattern
EXPLAIN ANALYZE SELECT * FROM files WHERE owner_id = '...';

-- Create index
CREATE INDEX CONCURRENTLY idx_files_owner ON files(owner_id);

-- For JSONB
CREATE INDEX CONCURRENTLY idx_files_metadata ON files USING GIN (metadata);
```

**Reindex Bloated:**
```sql
-- Check bloat
SELECT pg_relation_size('idx_files_metadata');

-- Reindex
REINDEX INDEX CONCURRENTLY idx_files_metadata;
```

## 5. Migration Management

### Version Control

All migrations are tracked in `schema_migrations` table:

```sql
SELECT version, applied_at 
FROM schema_migrations 
ORDER BY applied_at DESC;
```

### Migration Files

- `migrations/`
  - `001_initial_schema.sql` - Complete optimized schema
  - `001_create_users.up.sql` - Legacy incremental migrations
  - `...`
  - `023_upgrade_to_postgres_types.up.sql` - TEXT → JSONB/UUID

### Create New Migration

```bash
# Generate migration file
./scripts/create-migration.sh add_new_table

# Creates:
# - migrations/024_add_new_table.up.sql
# - migrations/024_add_new_table.down.sql
```

### Migration Best Practices

1. ✅ Always use `CONCURRENTLY` for index creation
2. ✅ Test in staging first
3. ✅ Include rollback (down.sql)
4. ✅ Add indexes for new columns
5. ✅ Update schema documentation
6. ✅ Run ANALYZE after migrations

## 6. Security Checklist

### Database Security

- [ ] Use SSL for connections
- [ ] Restrict connection sources (pg_hba.conf)
- [ ] Use connection pooler (pgBouncer)
- [ ] Enable row-level security (if multi-tenant)
- [ ] Encrypt sensitive data (pgcrypto)
- [ ] Regular security updates
- [ ] Audit logging enabled

### Access Control

```sql
-- Create application user with limited permissions
CREATE ROLE skeleton_app LOGIN PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE skeleton TO skeleton_app;
GRANT USAGE ON SCHEMA public TO skeleton_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO skeleton_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO skeleton_app;

-- For migrations
CREATE ROLE skeleton_admin LOGIN PASSWORD 'admin_password';
GRANT ALL ON DATABASE skeleton TO skeleton_admin;
GRANT ALL ON SCHEMA public TO skeleton_admin;
```

## 7. Troubleshooting

### Common Issues

**1. Slow Queries**
```sql
-- Find slow queries
SELECT query, calls, mean_time/1000 as avg_ms
FROM pg_stat_statements
ORDER BY mean_time DESC LIMIT 10;

-- Analyze specific query
EXPLAIN ANALYZE SELECT ...;
```

**2. Connection Issues**
```sql
-- Check connections
SELECT * FROM pg_stat_activity WHERE state = 'active';

-- Kill long-running query
SELECT pg_cancel_backend(pid);

-- Force disconnect
SELECT pg_terminate_backend(pid);
```

**3. Lock Wait Issues**
```sql
-- Find blocking queries
SELECT blocked.pid, blocking.pid, blocked.query
FROM pg_stat_activity blocked
JOIN pg_locks blocked_locks ON blocked.pid = blocked_locks.pid
JOIN pg_locks blocking_locks ON blocked_locks.locktype = blocking_locks.locktype
JOIN pg_stat_activity blocking ON blocking_locks.pid = blocking.pid
WHERE NOT blocked_locks.granted AND blocking_locks.granted;
```

**4. Disk Space Issues**
```sql
-- Find largest tables
SELECT tablename, pg_size_pretty(pg_total_relation_size(tablename))
FROM pg_tables
ORDER BY pg_total_relation_size(tablename) DESC;

-- Find bloated tables
VACUUM FULL tablename;
```

### Performance Tuning

**1. Configuration (postgresql.conf)**
```ini
# Memory
shared_buffers = 4GB
work_mem = 16MB
maintenance_work_mem = 1GB

# Connections
max_connections = 200

# Query Planner
random_page_cost = 1.1  # SSD
effective_cache_size = 12GB

# WAL
wal_buffers = 64MB
checkpoint_completion_target = 0.9

# Logging
log_min_duration_statement = 1000  # Log queries > 1s
log_checkpoints = on
log_connections = on
log_disconnections = on
```

**2. Autovacuum**
```ini
autovacuum = on
autovacuum_vacuum_scale_factor = 0.1
autovacuum_analyze_scale_factor = 0.05
autovacuum_vacuum_cost_limit = 1000
```

## 8. Backup & Recovery

### Backup Strategy

```bash
# Full backup (daily)
pg_dump $DATABASE_URL -Fc > backups/db_$(date +%Y%m%d).dump

# WAL archiving (continuous)
# postgresql.conf:
archive_mode = on
archive_command = 'cp %p /archive/%f'

# Point-in-time recovery
# Enable WAL archiving for recovery to any point in time
```

### Recovery Procedure

```bash
# Restore from backup
pg_restore -d skeleton_new backups/db_20260408.dump

# Point-in-time recovery
# 1. Stop PostgreSQL
# 2. Restore base backup
# 3. Configure recovery.conf
# 4. Start PostgreSQL
# 5. Monitor recovery progress
```

## 9. Scaling

### Vertical Scaling

- Increase shared_buffers
- Optimize work_mem
- Add CPU cores
- SSD storage

### Horizontal Scaling

- Read replicas
- Connection pooling
- Partitioning
- Sharding

### Partitioning Example

```sql
-- Time-based partitioning for audit logs
CREATE TABLE audit_records (
    id UUID,
    created_at TIMESTAMPTZ,
    ...
) PARTITION BY RANGE (created_at);

CREATE TABLE audit_records_2024_01 
    PARTITION OF audit_records 
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

## 10. Maintenance Schedule

### Daily
- [ ] Check slow query logs
- [ ] Monitor disk space
- [ ] Verify backups

### Weekly
- [ ] Run index analysis
- [ ] Check table bloat
- [ ] Review connection usage
- [ ] Refresh materialized views

### Monthly
- [ ] Full VACUUM ANALYZE
- [ ] Review query performance
- [ ] Update statistics
- [ ] Security audit

### Quarterly
- [ ] Capacity planning
- [ ] Index optimization review
- [ ] Configuration tuning
- [ ] Disaster recovery test

## Files Reference

| File | Purpose |
|------|---------|
| `scripts/deploy-staging.sh` | Deploy migrations to staging |
| `scripts/run-benchmarks.sh` | Run performance benchmarks |
| `scripts/optimize-indexes.sh` | Analyze and optimize indexes |
| `docs/monitoring/postgres_monitoring_queries.sql` | Monitoring queries |
| `docs/performance/BENCHMARKS.md` | Benchmark documentation |
| `POSTGRES_MIGRATION_SUMMARY.md` | Complete migration summary |

## Support

For issues or questions:
1. Check logs in `./logs/`
2. Review monitoring queries
3. Consult PostgreSQL documentation
4. Review migration files
5. Check benchmark results
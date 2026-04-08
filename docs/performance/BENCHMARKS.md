# PostgreSQL Performance Benchmarks

This directory contains benchmarks for PostgreSQL queries to ensure optimal performance.

## Running Benchmarks

```bash
# Run all benchmarks
make bench

# Run specific benchmark
go test -bench=BenchmarkGetFileByID -benchmem ./internal/benchmark

# Run with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./internal/benchmark
go tool pprof cpu.prof

# Run with memory profiling
go test -bench=. -memprofile=mem.prof ./internal/benchmark
go tool pprof mem.prof
```

## Benchmark Categories

### 1. File Repository Benchmarks

- **BenchmarkGetFileByID**: Single file lookup by primary key
- **BenchmarkListFilesByOwner**: Paginated file listing by owner
- **BenchmarkSearchFilesWithJSONB**: JSONB metadata query with GIN index

### 2. Task Queue Benchmarks

- **BenchmarkTaskQueue**: Worker fetching next task with `FOR UPDATE SKIP LOCKED`

### 3. Notification Benchmarks

- **BenchmarkNotificationBatch**: Batch notification creation

### 4. Audit Log Benchmarks

- **BenchmarkAuditLogInsert**: High-frequency audit log insertions

## Expected Performance (with indexes)

| Benchmark | Target Latency | Rows | Notes |
|-----------|---------------|------|-------|
| GetFileByID | < 1ms | 1 | Primary key lookup |
| ListFilesByOwner | < 5ms | 20 | Index on owner_id |
| SearchFilesWithJSONB | < 10ms | 20 | GIN index on metadata |
| TaskQueue | < 2ms | 1 | Complex with FOR UPDATE |
| NotificationBatch | < 50ms | 100 | Batch insert |
| AuditLogInsert | < 1ms | 1 | Simple insert |

## Index Verification

Before running benchmarks, verify indexes exist:

```sql
-- Check GIN indexes for JSONB
SELECT indexname, indexdef 
FROM pg_indexes 
WHERE indexdef LIKE '%GIN%';

-- Expected: idx_files_metadata_gin, idx_tasks_payload_gin, etc.

-- Check partial indexes
SELECT indexname, indexdef 
FROM pg_indexes 
WHERE indexdef LIKE '%WHERE%';

-- Check index usage
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;
```

## Performance Tips

### 1. JSONB Queries

Use containment operator `@>` for best GIN index performance:

```sql
-- GOOD: Uses GIN index
SELECT * FROM files WHERE metadata @> '{"width": 800}';

-- BAD: Cannot use GIN index efficiently
SELECT * FROM files WHERE metadata->>'width' = '800';
```

### 2. Task Queue

Use `FOR UPDATE SKIP LOCKED` to prevent lock contention:

```sql
-- GOOD: Skip locked rows, no waiting
SELECT id FROM tasks 
WHERE status = 'pending' 
FOR UPDATE SKIP LOCKED 
LIMIT 1;

-- BAD: Waits if row is locked
SELECT id FROM tasks 
WHERE status = 'pending' 
FOR UPDATE 
LIMIT 1;
```

### 3. Batch Inserts

Use batch operations instead of individual inserts:

```sql
-- GOOD: Single transaction with multiple values
INSERT INTO notifications (id, user_id, content)
VALUES 
  (uuid_generate_v7(), 'user1', 'content1'),
  (uuid_generate_v7(), 'user2', 'content2'),
  (uuid_generate_v7(), 'user3', 'content3');

-- BAD: Multiple transactions
INSERT INTO notifications (id, user_id, content) VALUES (uuid_generate_v7(), 'user1', 'content1');
INSERT INTO notifications (id, user_id, content) VALUES (uuid_generate_v7(), 'user2', 'content2');
```

### 4. Partial Indexes

Use partial indexes for common query patterns:

```sql
-- Index only active users
CREATE INDEX idx_users_active ON users (email) WHERE is_active = true;

-- Index only pending tasks
CREATE INDEX idx_tasks_pending ON tasks (scheduled_at) WHERE status = 'pending';
```

## Continuous Benchmarking

Add to CI/CD pipeline to detect performance regressions:

```yaml
# .github/workflows/benchmark.yml
- name: Run Benchmarks
  run: |
    go test -bench=. -benchmem ./internal/benchmark > benchmark_results.txt
    # Compare with baseline
    benchstat baseline.txt benchmark_results.txt
```

## Monitoring in Production

### Key Metrics to Monitor

```sql
-- Slow queries (> 100ms)
SELECT query, calls, total_time, mean_time, rows
FROM pg_stat_statements
WHERE mean_time > 100
ORDER BY mean_time DESC
LIMIT 10;

-- Index hit ratio (should be > 99%)
SELECT schemaname, tablename,
       idx_scan, idx_tup_read, idx_tup_fetch,
       idx_tup_fetch::float / NULLIF(idx_scan, 0) AS avg_tuples_per_scan
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Cache hit ratio (should be > 99%)
SELECT sum(heap_blks_read) AS heap_read,
       sum(heap_blks_hit) AS heap_hit,
       sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) AS ratio
FROM pg_statio_user_tables;

-- Table bloat (should be < 20%)
SELECT schemaname, tablename,
       n_dead_tup, n_live_tup,
       n_dead_tup::float / NULLIF(n_live_tup, 0) AS dead_ratio
FROM pg_stat_user_tables
WHERE n_live_tup > 1000
ORDER BY dead_ratio DESC;
```

### Regular Maintenance

```sql
-- Analyze tables to update statistics
ANALYZE files;
ANALYZE tasks;
ANALYZE notifications;

-- Vacuum tables to reclaim space
VACUUM ANALYZE files;

-- Reindex if bloat is high
REINDEX INDEX idx_files_metadata_gin;

-- Refresh materialized views
REFRESH MATERIALIZED VIEW user_activity_summary;
REFRESH MATERIALIZED VIEW file_storage_stats;
```
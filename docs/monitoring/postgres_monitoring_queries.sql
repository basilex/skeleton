-- PostgreSQL Monitoring Dashboard Queries
-- Use these queries in Grafana, DataDog, or your monitoring tool
-- File: docs/monitoring/postgres_monitoring_queries.sql

-- ============================================================================
--1. CONNECTION & PERFORMANCE OVERVIEW
-- ============================================================================

-- Active connections by state
SELECT 
    datname as database,
    usename as user,
    state,
    COUNT(*) as count,
    AVG(EXTRACT('epoch' FROM now() - query_start)) as avg_duration_seconds
FROM pg_stat_activity
WHERE datname IS NOT NULL
GROUP BY datname, usename, state
ORDER BY count DESC;

-- Long-running queries (> 5 seconds)
SELECT 
    pid,
    usename,
    datname,
    state,
    EXTRACT('epoch' FROM now() - query_start) as duration_seconds,
    query,
    query_start
FROM pg_stat_activity
WHERE state = 'active'
  AND now() - query_start > INTERVAL '5 seconds'
ORDER BY query_start ASC;

-- Blocked queries
SELECT 
    blocked.pid AS blocked_pid,
    blocked.usename AS blocked_user,
    blocking.pid AS blocking_pid,
    blocking.usename AS blocking_user,
    blocked.query AS blocked_query,
    blocking.query AS blocking_query
FROM pg_stat_activity blocked
JOIN pg_locks blocked_locks ON blocked.pid = blocked_locks.pid
JOIN pg_locks blocking_locks ON blocked_locks.locktype = blocking_locks.locktype
    AND blocked_locks.database IS NOT DISTINCT FROM blocking_locks.database
    AND blocked_locks.relation IS NOT DISTINCT FROM blocking_locks.relation
    AND blocked_locks.page IS NOT DISTINCT FROM blocking_locks.page
    AND blocked_locks.tuple IS NOT DISTINCT FROM blocking_locks.tuple
    AND blocked_locks.pid != blocking_locks.pid
JOIN pg_stat_activity blocking ON blocking_locks.pid = blocking.pid
WHERE NOT blocked_locks.granted;

-- ============================================================================
-- 2. INDEX HEALTH & PERFORMANCE
-- ============================================================================

-- Index usage statistics
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
    CASE 
        WHEN idx_scan = 0 THEN 'UNUSED'
        WHEN idx_scan < 100 THEN 'LOW_USAGE'
        ELSE 'ACTIVE'
    END as status
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Unused indexes (candidates for removal)
SELECT 
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
    idx_scan as scans,
    'DROP INDEX ' || schemaname || '.' || indexname || ';' as drop_command
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexname NOT LIKE '%_pkey'
ORDER BY pg_relation_size(indexrelid) DESC;

-- Missing indexes (sequential scans > 1000)
SELECT 
    schemaname,
    tablename,
    seq_scan as sequential_scans,
    idx_scan as index_scans,
    seq_tup_read as seq_tuples_read,
    CASE 
        WHEN seq_scan > 1000 AND idx_scan = 0 THEN 'CRITICAL'
        WHEN seq_scan > 1000 AND idx_scan < seq_scan::float * 0.1 THEN 'WARNING'
        ELSE 'OK'
    END as status
FROM pg_stat_user_tables
WHERE seq_scan > 1000
ORDER BY seq_scan DESC;

-- Index hit ratio (should be > 99%)
SELECT 
    schemaname,
    tablename,
    round(idx_tup_read::float / NULLIF(idx_scan, 0), 2) as avg_tuples_per_scan,
    round((idx_tup_read::float / NULLIF(idx_scan + seq_scan, 0)) * 100, 2) as index_hit_ratio_pct,
    CASE 
        WHEN idx_scan = 0 THEN 'NO_INDEX_USE'
        WHEN (idx_tup_read::float / NULLIF(idx_scan + seq_scan, 0)) < 0.95 THEN 'POOR'
        WHEN (idx_tup_read::float / NULLIF(idx_scan + seq_scan, 0)) < 0.99 THEN 'FAIR'
        ELSE 'GOOD'
    END as status
FROM pg_stat_user_tables
WHERE idx_scan + seq_scan > 100
ORDER BY index_hit_ratio_pct ASC;

-- ============================================================================
-- 3. QUERY PERFORMANCE
-- ============================================================================

-- Slow queries from pg_stat_statements (requires extension)
SELECT 
    calls,
    round(total_time::float / 1000, 2) as total_time_seconds,
    round(mean_time::float, 2) as mean_time_ms,
    round((100 * total_time / SUM(total_time) OVER ())::float, 2) as percent_total,
    rows,
    query
FROM pg_stat_statements
ORDER BY total_time DESC
LIMIT 20;

-- Query performance by type
SELECT 
    CASE 
        WHEN query LIKE 'SELECT%' THEN 'SELECT'
        WHEN query LIKE 'INSERT%' THEN 'INSERT'
        WHEN query LIKE 'UPDATE%' THEN 'UPDATE'
        WHEN query LIKE 'DELETE%' THEN 'DELETE'
        ELSE 'OTHER'
    END as query_type,
    COUNT(*) as count,
    AVG(mean_time) as avg_time_ms,
    MAX(mean_time) as max_time_ms,
    SUM(total_time) as total_time_seconds
FROM pg_stat_statements
GROUP BY query_type
ORDER BY total_time_seconds DESC;

-- ============================================================================
-- 4. TABLE BLOAT & MAINTENANCE
-- ============================================================================

-- Table bloat (dead tuples)
SELECT 
    schemaname,
    tablename,
    n_live_tup as live_tuples,
    n_dead_tup as dead_tuples,
    round(n_dead_tup::float / NULLIF(n_live_tup, 0) * 100, 2) as bloat_ratio_pct,
    CASE 
        WHEN n_dead_tup::float / NULLIF(n_live_tup, 0) > 0.2 THEN 'VACUUM_NEEDED'
        WHEN n_dead_tup::float / NULLIF(n_live_tup, 0) > 0.1 THEN 'WARNING'
        ELSE 'OK'
    END as status,
    'VACUUM ANALYZE ' || schemaname || '.' || tablename || ';' as vacuum_command
FROM pg_stat_user_tables
WHERE n_live_tup > 1000
ORDER BY bloat_ratio_pct DESC;

-- Last vacuum/analyze times
SELECT 
    schemaname,
    tablename,
    last_vacuum,
    last_autovacuum,
    last_analyze,
    last_autoanalyze,
    CASE 
        WHEN last_vacuum IS NULL AND last_autovacuum IS NULL THEN 'NEVER_VACUUMED'
        WHEN last_vacuum < NOW() - INTERVAL '7 days' AND last_autovacuum < NOW() - INTERVAL '7 days' THEN 'OLD'
        ELSE 'RECENT'
    END as vacuum_status
FROM pg_stat_user_tables
ORDER BY last_vacuum ASC NULLS FIRST;

-- ============================================================================
-- 5. CACHE & BUFFER PERFORMANCE
-- ============================================================================

-- Buffer cache hit ratio (should be > 99%)
SELECT 
    'Index cache hit ratio' as metric,
    round((sum(idx_blks_hit)::float / (sum(idx_blks_hit) + sum(idx_blks_read))) * 100, 2) as value_pct,
    CASE 
        WHEN sum(idx_blks_hit)::float / (sum(idx_blks_hit) + sum(idx_blks_read)) < 0.95 THEN 'CRITICAL'
        WHEN sum(idx_blks_hit)::float / (sum(idx_blks_hit) + sum(idx_blks_read)) < 0.99 THEN 'WARNING'
        ELSE 'GOOD'
    END as status
FROM pg_statio_user_indexes
UNION ALL
SELECT 
    'Table cache hit ratio' as metric,
    round((sum(heap_blks_hit)::float / (sum(heap_blks_hit) + sum(heap_blks_read))) * 100, 2) as value_pct,
    CASE 
        WHEN sum(heap_blks_hit)::float / (sum(heap_blks_hit) + sum(heap_blks_read)) < 0.95 THEN 'CRITICAL'
        WHEN sum(heap_blks_hit)::float / (sum(heap_blks_hit) + sum(heap_blks_read)) < 0.99 THEN 'WARNING'
        ELSE 'GOOD'
    END as status
FROM pg_statio_user_tables;

-- ============================================================================
-- 6. STORAGE & SIZE ANALYSIS
-- ============================================================================

-- Top 10 largest tables
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
    pg_size_pretty(pg_indexes_size(schemaname||'.'||tablename)) as indexes_size,
    n_live_tup as row_count
FROM pg_stat_user_tables
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
LIMIT 10;

-- Database size over time (for trending)
SELECT 
    datname,
    pg_size_pretty(pg_database_size(datname)) as size,
    pg_database_size(datname) as bytes
FROM pg_database
WHERE datname = current_database();

-- Index size vs table size ratio
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
    pg_size_pretty(pg_indexes_size(schemaname||'.'||tablename)) as indexes_size,
    round(pg_indexes_size(schemaname||'.'||tablename)::float / NULLIF(pg_relation_size(schemaname||'.'||tablename), 0) * 100, 2) as index_ratio_pct,
    CASE 
        WHEN pg_indexes_size(schemaname||'.'||tablename)::float / pg_relation_size(schemaname||'.'||tablename) > 0.5 THEN 'HIGH'
        WHEN pg_indexes_size(schemaname||'.'||tablename)::float / pg_relation_size(schemaname||'.'||tablename) > 0.3 THEN 'MODERATE'
        ELSE 'LOW'
    END as status
FROM pg_stat_user_tables
WHERE pg_relation_size(schemaname||'.'||tablename) > 0
ORDER BY index_ratio_pct DESC;

-- ============================================================================
-- 7. LOCK & WAIT ANALYSIS
-- ============================================================================

-- Current locks
SELECT 
    datname,
    usename,
    pid,
    mode,
    locktype,
    granted,
    query_start,
    EXTRACT('epoch' FROM now() - query_start) as age_seconds
FROM pg_locks l
JOIN pg_stat_activity a ON l.pid = a.pid
WHERE NOT l.granted
ORDER BY query_start;

-- Waiting events (PostgreSQL 10+)
SELECT 
    event_type,
    event,
    COUNT(*) as waiters
FROM pg_stat_activity
CROSS JOIN LATERAL unnest(wait_event_type, wait_event) AS t(event_type, event)
WHERE state = 'active'
  AND wait_event_type IS NOT NULL
GROUP BY event_type, event
ORDER BY waiters DESC;

-- ============================================================================
-- 8. JSONB & SPECIALIZED QUERIES
-- ============================================================================

-- JSONB index usage
SELECT 
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as size,
    idx_scan,
    CASE 
        WHEN indexdef LIKE '%USING gin%' THEN 'GIN'
        WHEN indexdef LIKE '%USING btree%' THEN 'BTREE'
        ELSE 'OTHER'
    END as index_type
FROM pg_stat_user_indexes
JOIN pg_indexes ON pg_stat_user_indexes.indexrelname = pg_indexes.indexname
WHERE indexname LIKE '%metadata%' OR indexname LIKE '%payload%' OR indexname LIKE '%details%'
ORDER BY idx_scan DESC;

-- JSONB query performance examples
EXPLAIN ANALYZE SELECT * FROM files WHERE metadata @> '{"width": 800}';
EXPLAIN ANALYZE SELECT * FROM tasks WHERE payload @> '{"type": "email"}';

-- ============================================================================
-- 9. REPLICATION & WAL (if applicable)
-- ============================================================================

-- WAL position (for replication lag)
SELECT 
    pg_current_wal_lsn() as current_wal_position,
    pg_wal_lsn_diff(pg_current_wal_lsn(), '0/0') as wal_bytes;

-- Replication status (if replica exists)
SELECT 
    client_addr,
    state,
    sync_state,
    sent_lsn,
    write_lsn,
    flush_lsn,
    replay_lsn,
    EXTRACT('epoch' FROM (now() - reply_date)) as lag_seconds
FROM pg_stat_replication;

-- ============================================================================
-- 10. SCHEDULED MAINTENANCE QUERIES
-- ============================================================================

-- Refresh materialized views
REFRESH MATERIALIZED VIEW user_activity_summary;
REFRESH MATERIALIZED VIEW file_storage_stats;

-- Analyze tables for query planner
ANALYZE files;
ANALYZE tasks;
ANALYZE notifications;
ANALYZE users;

-- Reindex if bloat is high
REINDEX INDEX idx_files_metadata;
REINDEX INDEX idx_tasks_payload;

-- Clean expired files (if function exists)
SELECT clean_expired_files();

-- Mark stalled tasks as failed (if function exists)
SELECT mark_stalled_tasks_failed(INTERVAL '30 minutes');
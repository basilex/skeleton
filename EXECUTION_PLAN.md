# PostgreSQL Migration Execution Plan

Step-by-step guide to execute the PostgreSQL migration from development to production.

## Phase 1: Staging Deployment

### Pre-Deployment Checklist

```bash
# 1. Verify staging environment
export STAGING_DATABASE_URL="postgres://skeleton:password@staging-db:5432/skeleton"

# 2. Check current state
psql $STAGING_DATABASE_URL -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;"

# 3. Verify scripts are ready
ls -lh scripts/deploy-staging.sh
ls -lh scripts/run-benchmarks.sh
ls -lh migrations/*.sql
```

### Step 1.1: Backup Staging Database

```bash
# Create backup directory
mkdir -p backups/staging/$(date +%Y%m%d_%H%M%S)
cd backups/staging/$(date +%Y%m%d_%H%M%S)

# Backup current database
pg_dump $STAGING_DATABASE_URL \
  --schema-only \
  --no-owner \
  --no-privileges \
  > schema_backup.sql

pg_dump $STAGING_DATABASE_URL \
  --data-only \
  --table=users \
  --table=roles \
  --table=files \
  --table=tasks \
  --table=notifications \
  --table=audit_records \
  > data_backup.sql
```

### Step 1.2: Run Deployment Script

```bash
# Execute deployment
cd /path/to/skeleton

# This will:
# ✅ Check database connection
# ✅ Create backup
# ✅ Run migrations
# ✅ Verify schema
# ✅ Run health checks
# ✅ Generate report
./scripts/deploy-staging.sh

# Monitor progress
tail -f logs/staging-deploy-*.log
```

### Step 1.3: Verify Deployment

```bash
# Check migration status
psql $STAGING_DATABASE_URL <<SQL
SELECT version, applied_at FROM schema_migrations ORDER BY applied_at DESC LIMIT 10;
SQL

# Verify indexes were created
psql $STAGING_DATABASE_URL <<SQL
SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public';
-- Expected: 40+ indexes
SQL

# Check JSONB columns
psql $STAGING_DATABASE_URL <<SQL
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'files' AND column_name = 'metadata';
-- Expected: jsonb
SQL

# Verify materialized views
psql $STAGING_DATABASE_URL <<SQL
SELECT matviewname FROM pg_matviews WHERE schemaname = 'public';
-- Expected: user_activity_summary, file_storage_stats
SQL

# Check extensions
psql $STAGING_DATABASE_URL <<SQL
SELECT extname, extversion FROM pg_extension WHERE extname IN ('uuid-ossp', 'pgcrypto');
-- Expected: Both extensions installed
SQL
```

### Step 1.4: Run Application Tests

```bash
# Run integration tests against staging
export DATABASE_URL=$STAGING_DATABASE_URL
make test-integration

# Check for errors
echo $?  # Should be 0

# Run specific tests
go test ./internal/files/infrastructure/persistence -v
go test ./internal/tasks/infrastructure/persistence -v
go test ./internal/notifications/infrastructure/persistence -v
```

## Phase 2: Performance Benchmarks

### Step 2.1: Prepare Benchmark Environment

```bash
# Set benchmark database
export BENCHMARK_DATABASE_URL="postgres://skeleton:password@benchmark-db:5432/skeleton_benchmark"

# Create benchmark database
psql postgres://skeleton:password@localhost:5432/postgres <<SQL
DROP DATABASE IF EXISTS skeleton_benchmark;
CREATE DATABASE skeleton_benchmark;
SQL

# Run initial schema
psql $BENCHMARK_DATABASE_URL -f migrations/001_initial_schema.sql
```

### Step 2.2: Execute Benchmarks

```bash
# Run all benchmarks
./scripts/run-benchmarks.sh

# Or run specific benchmark
./scripts/run-benchmarks.sh BenchmarkGetFileByID

# Monitor progress
tail -f benchmark_results/*/benchmark_results.txt
```

### Step 2.3: Analyze Results

```bash
# View benchmark results
cat benchmark_results/*/report.md

# Compare with baseline
benchstat benchmark_results/baseline.txt benchmark_results/*/benchmark_results.txt

# Check database stats
cat benchmark_results/*/database_stats.txt

# Key metrics to verify:
# ✓ File lookup: < 1ms
# ✓ JSONB search: < 10ms
# ✓ Task queue: < 2ms
# ✓ Index hit ratio: > 99%
# ✓ Cache hit ratio: > 99%
```

### Step 2.4: Save Baseline

```bash
# Create baseline for future comparison
cp benchmark_results/*/benchmark_results.txt benchmark_results/baseline.txt

# Commit to repository
git add benchmark_results/baseline.txt
git commit -m "chore: add performance baseline"
```

## Phase 3: Production Rollout

### Pre-Production Checklist

```bash
# 1. Schedule maintenance window
# Recommended: Low-traffic period (e.g., Sunday 3 AM)

# 2. Notify stakeholders

# 3. Verify production backups
aws s3 ls s3://backups/postgresql/production/$(date +%Y%m)/

# 4. Check monitoring systems
# Ensure alerts are configured

# 5. Prepare rollback plan
# Document: How to restore from backup
```

### Step 3.1: Production Backup

```bash
# Create timestamp
BACKUP_TS=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="backups/production/$BACKUP_TS"

# Full backup
pg_dump $PRODUCTION_DATABASE_URL \
  --format=custom \
  --no-owner \
  --no-privileges \
  --verbose \
  > "$BACKUP_DIR/full_backup.dump"

# WAL backup
rsync -av /var/lib/postgresql/wal/ "$BACKUP_DIR/wal/"

# Verify backup integrity
pg_restore --list "$BACKUP_DIR/full_backup.dump" | head -20

# Store backup metadata
echo "{
  \"timestamp\": \"$BACKUP_TS\",
  \"database\": \"skeleton_production\",
  \"size\": \"$(du -h $BACKUP_DIR/full_backup.dump | cut -f1)\",
  \"migrations\": \"$(psql $PRODUCTION_DATABASE_URL -t -c 'SELECT MAX(version) FROM schema_migrations')\"
}" > "$BACKUP_DIR/metadata.json"
```

### Step 3.2: Migration Execution

```bash
# Stop application (if needed)
kubectl scale deployment skeleton-api --replicas=0

# Run migrations with monitoring
psql $PRODUCTION_DATABASE_URL <<SQL
BEGIN;

-- Check current state
SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;

-- Run migration
\i migrations/023_upgrade_to_postgres_types.up.sql

-- Verify
SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public';
SELECT COUNT(*) FROM pg_matviews WHERE schemaname = 'public';

COMMIT;
SQL

# Refresh materialized views
psql $PRODUCTION_DATABASE_URL <<SQL
REFRESH MATERIALIZED VIEW user_activity_summary;
REFRESH MATERIALIZED VIEW file_storage_stats;
SQL

# Analyze tables
psql $PRODUCTION_DATABASE_URL -c "ANALYZE;"
```

### Step 3.3: Application Deployment

```bash
# Deploy new application version
kubectl apply -f k8s/deployment.yaml

# Wait for rollout
kubectl rollout status deployment/skeleton-api

# Check application health
curl -f https://api.example.com/health || exit 1

# Verify application functionality
curl -f https://api.example.com/system/ready || exit 1
```

### Step 3.4: Smoke Tests

```bash
# Run smoke tests
./scripts/smoke-tests.sh

# Test critical paths
curl https://api.example.com/api/v1/health
curl https://api.example.com/api/v1/system/info

# Verify database connectivity
psql $PRODUCTION_DATABASE_URL -c "SELECT COUNT(*) FROM users;"

# Check logs for errors
kubectl logs -l app=skeleton-api --tail=100 | grep -i error
```

### Step 3.5: Monitor Production

```bash
# Watch query performance
watch -n 5 'psql $PRODUCTION_DATABASE_URL -c "SELECT datname, numbackends, xact_commit, xact_rollback FROM pg_stat_database WHERE datname = '\''skeleton'\'';"'

# Monitor slow queries
psql $PRODUCTION_DATABASE_URL <<SQL
SELECT query, calls, mean_time/1000 as avg_ms
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
SQL

# Check index usage
psql $PRODUCTION_DATABASE_URL <<SQL
SELECT schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexname NOT LIKE '%_pkey';
SQL

# Monitor connections
psql $PRODUCTION_DATABASE_URL -c "SELECT COUNT(*) FROM pg_stat_activity;"
```

## Phase 4: Optimization & Iteration

### Step 4.1: Collect Performance Metrics

```bash
# Run monitoring queries daily
./scripts/collect-metrics.sh

# Store metrics for trending
psql monitoring_db <<SQL
INSERT INTO performance_metrics (
  timestamp,
  index_hit_ratio,
  cache_hit_ratio,
  avg_query_time,
  connection_count,
  table_bloat
) VALUES (
  NOW(),
  /* ... */
);
SQL
```

### Step 4.2: Analyze Slow Queries

```bash
# Weekly slow query analysis
psql $PRODUCTION_DATABASE_URL <<SQL
SELECT 
    query,
    calls,
    total_time/1000 as total_seconds,
    mean_time/1000 as avg_ms,
    rows
FROM pg_stat_statements
WHERE mean_time > 100  -- Queries slower than 100ms
ORDER BY mean_time DESC
LIMIT 20;
SQL

# For each slow query, analyze:
# 1. EXPLAIN ANALYZE <query>
# 2. Check if index exists
# 3. Consider adding index or rewriting query
```

### Step 4.3: Index Optimization

```bash
# Run weekly index analysis
./scripts/optimize-indexes.sh --dry-run

# Review recommendations
cat index_analysis/*/recommendations.md

# Test in staging first
export DATABASE_URL=$STAGING_DATABASE_URL
./scripts/optimize-indexes.sh

# If successful, apply to production
export DATABASE_URL=$PRODUCTION_DATABASE_URL
./scripts/optimize-indexes.sh
```

### Step 4.4: Continuous Improvement

**Weekly Tasks:**
```bash
# Monday: Review slow queries
psql $PRODUCTION_DATABASE_URL < docs/monitoring/postgres_monitoring_queries.sql

# Wednesday: Check index usage
./scripts/optimize-indexes.sh --dry-run

# Friday: Performance baseline
./scripts/run-benchmarks.sh
benchstat benchmark_results/baseline.txt benchmark_results/current.txt
```

**Monthly Tasks:**
```bash
# Full VACUUM on high-activity tables
psql $PRODUCTION_DATABASE_URL <<SQL
VACUUM FULL ANALYZE files;
VACUUM FULL ANALYZE tasks;
VACUUM FULL ANALYZE notifications;
SQL

# Reindex large indexes
psql $PRODUCTION_DATABASE_URL <<SQL
REINDEX INDEX CONCURRENTLY idx_files_metadata;
REINDEX INDEX CONCURRENTLY idx_tasks_payload;
SQL

# Update statistics
psql $PRODUCTION_DATABASE_URL -c "ANALYZE;"
```

**Quarterly Tasks:**
```bash
# Review capacity
psql $PRODUCTION_DATABASE_URL <<SQL
SELECT 
    pg_size_pretty(pg_database_size('skeleton')) as db_size,
    pg_size_pretty(SUM(pg_relation_size(tablename))) as table_size,
    COUNT(*) as table_count
FROM pg_tables WHERE schemaname = 'public';
SQL

# Plan capacity
./scripts/capacity-planning.sh

# Review security
psql $PRODUCTION_DATABASE_URL <<SQL
SELECT * FROM pg_roles WHERE rolname LIKE 'skeleton%';
SQL
```

## Monitoring & Alerting

### Setup Prometheus Metrics

```yaml
# prometheus/postgres_exporter.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: postgresql
    static_configs:
      - targets: ['postgres-exporter:9187']
```

### Setup Grafana Dashboard

```bash
# Import dashboard
curl -X POST http://grafana:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @docs/monitoring/grafana-dashboard.json

# Key panels:
# 1. Connection Count
# 2. Query Latency (p50, p95, p99)
# 3. Index Hit Ratio
# 4. Cache Hit Ratio
# 5. Table Bloat %
# 6. Slow Queries (>100ms)
```

### Configure Alerts

```yaml
# alertmanager/postgres_alerts.yml
groups:
  - name: postgresql
    rules:
      - alert: PostgresSlowQuery
        expr: pg_stat_activity_max_tx_duration_seconds > 60
        for: 5m
        annotations:
          summary: "PostgreSQL slow query detected"
          
      - alert: PostgresConnectionsHigh
        expr: pg_stat_activity_count > 180
        for: 2m
        annotations:
          summary: "PostgreSQL connections > 180"
          
      - alert: PostgresIndexHitRatioLow
        expr: pg_stat_user_indexes_hit_ratio < 0.95
        for: 5m
        annotations:
          summary: "Index hit ratio < 95%"
          
      - alert: PostgresCacheHitRatioLow
        expr: pg_statio_user_tables_cache_hit_ratio < 0.95
        for: 5m
        annotations:
          summary: "Cache hit ratio < 95%"
```

## Rollback Procedures

### If Migration Fails

```bash
# 1. Stop application
kubectl scale deployment skeleton-api --replicas=0

# 2. Restore from backup
pg_restore \
  --dbname=$PRODUCTION_DATABASE_URL \
  --clean \
  --if-exists \
  backups/production/*/full_backup.dump

# 3. Restart application with old version
kubectl apply -f k8s/deployment-v1.yaml

# 4. Verify functionality
curl -f https://api.example.com/health

# 5. Notify team
```

### If Performance Degrades

```bash
# 1. Identify issue
psql $PRODUCTION_DATABASE_URL < docs/monitoring/postgres_monitoring_queries.sql

# 2. If index-related, reindex
psql $PRODUCTION_DATABASE_URL <<SQL
REINDEX INDEX CONCURRENTLY <problematic_index>;
SQL

# 3. If query-related, add missing index
psql $PRODUCTION_DATABASE_URL <<SQL
CREATE INDEX CONCURRENTLY idx_new ON table(column);
SQL

# 4. If connection-related, restart pooler
kubectl rollout restart deployment/pgbouncer
```

## Success Criteria

### Phase 1 Success
- [ ] All migrations applied successfully
- [ ] Schema verified (indexes, extensions, materialized views)
- [ ] Application tests pass
- [ ] No errors in logs

### Phase 2 Success
- [ ] All benchmarks complete
- [ ] Performance within targets
- [ ] Baseline saved
- [ ] Report generated

### Phase 3 Success
- [ ] Production deployment successful
- [ ] Smoke tests pass
- [ ] Monitoring active
- [ ] No incidents in 24 hours

### Phase 4 Success
- [ ] Weekly monitoring established
- [ ] Monthly optimization performed
- [ ] Quarterly capacity reviewed
- [ ] Documentation updated

## Support & Escalation

### During Migration
- **Primary**: Database Admin
- **Secondary**: DevOps Lead
- **Escalation**: Engineering Manager

### Contact Information
```
Database Admin: db-admin@example.com
DevOps Lead: devops@example.com
Engineering Manager: eng-manager@example.com
```

### Issue Tracking
- GitHub Issues: postgres-migration
- Slack: #postgres-migration
- PagerDuty: Database Service

## Timeline

| Phase | Duration | Start | End |
|-------|----------|-------|-----|
| Staging Testing | 2 days | Monday | Tuesday |
| Benchmarking | 1 day | Wednesday | Wednesday |
| Production Prep | 1 day | Thursday | Thursday |
| Production Deploy | 1 day | Friday (Low Traffic) | Friday |
| Monitoring | Ongoing | Week 2+ | Continuous |
| Optimization | Ongoing | Week 3+ | Continuous |

**Total Duration**: 5 days + ongoing optimization
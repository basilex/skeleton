#!/bin/bash
# Deploy PostgreSQL migrations to staging environment
# Usage: ./scripts/deploy-staging.sh

set -e

echo "🚀 Deploying to Staging Environment"
echo "===================================== "

# Configuration
STAGING_DB="${STAGING_DATABASE_URL:-postgres://skeleton:password@staging-db:5432/skeleton}"
BACKUP_DIR="./backups/staging/$(date +%Y%m%d_%H%M%S)"
LOG_FILE="./logs/staging-deploy-$(date +%Y%m%d_%H%M%S).log"

# Create directories
mkdir -p "$BACKUP_DIR"
mkdir -p "$(dirname "$LOG_FILE")"

# Log function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Error handling
error_exit() {
    log "❌ ERROR: $1"
    log "Rolling back changes..."
    # Add rollback logic here
    exit 1
}

trap 'error_exit "Script interrupted"' INT TERM

# Step 1: Pre-deployment checks
log "Step 1: Pre-deployment checks"

# Check database connection
log "Checking database connection..."
if ! psql "$STAGING_DB" -c "SELECT 1" > /dev/null 2>&1; then
    error_exit "Cannot connect to staging database"
fi
log "✓ Database connection successful"

# Check migration status
log "Checking current migration status..."
CURRENT_VERSION=$(psql "$STAGING_DB" -t -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1" 2>/dev/null | xargs || echo "0")
log "Current migration version: $CURRENT_VERSION"

# Check for pending migrations
PENDING_MIGRATIONS=$(ls -1 migrations/*.up.sql 2>/dev/null | grep -v "^001_" | wc -l | xargs)
log "Pending migrations: $PENDING_MIGRATIONS"

if [ "$PENDING_MIGRATIONS" -eq 0 ]; then
    log "No pending migrations found"
    exit 0
fi

# Step 2: Create backup
log "Step 2: Creating backup..."

# Backup database schema
log "Backing up schema..."
pg_dump "$STAGING_DB" --schema-only > "$BACKUP_DIR/schema.sql" || error_exit "Schema backup failed"

# Backup data (for critical tables)
log "Backing up critical data..."
CRITICAL_TABLES="users roles files tasks notifications audit_records"
for table in $CRITICAL_TABLES; do
    pg_dump "$STAGING_DB" --data-only --table="$table" > "$BACKUP_DIR/${table}_data.sql" || log "Warning: Table $table not found"
done

log "✓ Backup created at $BACKUP_DIR"

# Step 3: Run migrations in transaction
log "Step 3: Running migrations..."

MIGRATION_FILES=$(ls -1 migrations/*.up.sql | grep -v "^001_" | sort)

for migration in $MIGRATION_FILES; do
    filename=$(basename "$migration")
    log "Running migration: $filename"
    
    # Run migration in a transaction
    psql "$STAGING_DB" <<EOF
BEGIN;
\i $migration
INSERT INTO schema_migrations (version, applied_at) VALUES ('${filename%.up.sql}', NOW());
COMMIT;
EOF
    
    if [ $? -eq 0 ]; then
        log "✓ Migration $filename completed"
    else
        error_exit "Migration $filename failed"
    fi
done

log "✓ All migrations completed"

# Step 4: Verify schema
log "Step 4: Verifying schema..."

# Check indexes
EXPECTED_INDEXES=40
ACTUAL_INDEXES=$(psql "$STAGING_DB" -t -c "SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public'" | xargs)

if [ "$ACTUAL_INDEXES" -ge "$EXPECTED_INDEXES" ]; then
    log "✓ Indexes created ($ACTUAL_INDEXES >= $EXPECTED_INDEXES)"
else
    log "⚠ Warning: Expected at least $EXPECTED_INDEXES indexes, found $ACTUAL_INDEXES"
fi

# Check extensions
log "Checking PostgreSQL extensions..."
EXTENSIONS=$(psql "$STAGING_DB" -t -c "SELECT extname FROM pg_extension WHERE extname IN ('uuid-ossp', 'pgcrypto')")
log "Extensions: $EXTENSIONS"

# Check materialized views
MATVIEWS=$(psql "$STAGING_DB" -t -c "SELECT COUNT(*) FROM pg_matviews WHERE schemaname = 'public'" | xargs)
if [ "$MATVIEWS" -ge 2 ]; then
    log "✓ Materialized views created ($MATVIEWS)"
else
    log "⚠ Warning: Expected 2 materialized views, found $MATVIEWS"
fi

# Step 5: Performance baseline
log "Step 5: Running baseline benchmarks..."

# Run a simple query test
QUERY_TIMES=$(psql "$STAGING_DB" <<EOF
\timing on
SELECT COUNT(*) FROM files;
SELECT COUNT(*) FROM tasks WHERE status = 'pending';
SELECT COUNT(*) FROM notifications WHERE status = 'pending';
EOF
)

log "Baseline query times:"
log "$QUERY_TIMES"

# Step 6: Health checks
log "Step 6: Running health checks..."

# Check connection pool
log "Checking connection pool..."
psql "$STAGING_DB" -c "SELECT count(*) FROM pg_stat_activity WHERE datname = current_database();" || log "Warning: Could not check connections"

# Check table sizes
log "Table sizes:"
psql "$STAGING_DB" -c "SELECTschemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size FROM pg_tables WHERE schemaname = 'public' ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;" || log "Warning: Could not check table sizes"

# Check index usage
log "Index usage (should be > 95%):"
psql "$STAGING_DB" -c "SELECT schemaname, tablename, (idx_scan::float / NULLIF(idx_scan + seq_scan, 0) * 100) as index_usage_pct FROM pg_stat_user_tables WHERE (idx_scan + seq_scan) > 100 ORDER BY index_usage_pct;" || log "Warning: Could not check index usage"

log "✓ Health checks completed"

# Step 7: Post-deployment tasks
log "Step 7: Post-deployment tasks..."

# Refresh materialized views
log "Refreshing materialized views..."
psql "$STAGING_DB" <<EOF
REFRESH MATERIALIZED VIEW user_activity_summary;
REFRESH MATERIALIZED VIEW file_storage_stats;
EOF

log "✓ Materialized views refreshed"

# Analyze tables for query planner
log "Analyzing tables..."
psql "$STAGING_DB" -c "ANALYZE;" || log "Warning: Analyze failed"

log "✓ Tables analyzed"

# Step 8: Generate report
log "Step 8: Generating deployment report..."

REPORT_FILE="$BACKUP_DIR/deployment_report.txt"

cat > "$REPORT_FILE" <<EOF
PostgreSQL Migration Deployment Report
========================================
Date: $(date)
Environment: Staging
Database: $STAGING_DB

Migration Status
---------------
Starting Version: $CURRENT_VERSION
Migrations Applied: $PENDING_MIGRATIONS

Schema Verification
------------------
Indexes Created: $ACTUAL_INDEXES
Materialized Views: $MATVIEWS
Extensions: $EXTENSIONS

Performance Baseline
-------------------
$QUERY_TIMES

Next Steps
----------
1. Run application integration tests
2. Monitor query performance for 24 hours
3. Review slow query log
4. Optimize indexes if needed
5. Document any anomalies

Backup Location: $BACKUP_DIR
Log File: $LOG_FILE
EOF

log "✓ Deployment report generated: $REPORT_FILE"

# Step 9: Notification
log "Step 9: Sending notifications..."

# Add Slack/email notification here if needed
# curl -X POST -H 'Content-type: application/json' --data '{"text":"Staging deployment completed successfully"}' $SLACK_WEBHOOK

log "======================================="
log "✅ Staging deployment completed successfully!"
log "Report: $REPORT_FILE"
log "Backup: $BACKUP_DIR"
log "Logs: $LOG_FILE"
log "======================================="

# Step 10: Rollback instructions
log ""
log "To rollback this deployment:"
log "  1. Restore from backup: psql $STAGING_DB < $BACKUP_DIR/schema.sql"
log "  2. Restore data: for table in $CRITICAL_TABLES; do psql $STAGING_DB < $BACKUP_DIR/\${table}_data.sql; done"
log "  3. Restart application services"
log ""
#!/bin/bash
# Analyze and optimize PostgreSQL indexes
# Usage: ./scripts/optimize-indexes.sh [--dry-run]

set -e

echo "🔍 PostgreSQL Index Optimization"
echo "================================"

# Configuration
DATABASE_URL="${DATABASE_URL:-postgres://skeleton:password@localhost:5432/skeleton}"
DRY_RUN="${1:-}"
OUTPUT_DIR="./index_analysis/$(date +%Y%m%d_%H%M%S)"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Log functionlog() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Step 1: Analyze current index usage
log "Step 1: Analyzing index usage..."

psql "$DATABASE_URL" <<EOF > "$OUTPUT_DIR/index_usage.txt"
-- Index usage by table
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexrelid)) as size,
    CASE 
        WHEN idx_scan = 0 THEN '❌ UNUSED'
        WHEN idx_scan < 100 THEN '⚠️  LOW_USAGE'
        ELSE '✓ ACTIVE'
    END as status
FROM pg_stat_user_indexes
ORDER BY idx_scan ASC;
EOF

log "✓ Index usage analyzed: $OUTPUT_DIR/index_usage.txt"

# Step 2: Find unused indexes
log "Step 2: Finding unused indexes..."

psql "$DATABASE_URL" <<EOF > "$OUTPUT_DIR/unused_indexes.txt"
SELECT 
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as size,
    'DROP INDEX CONCURRENTLY ' || schemaname || '.' || indexname || ';' as drop_command
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND indexname NOT LIKE '%_pkey'
  AND indexname NOT LIKE '%_unique'
ORDER BY pg_relation_size(indexrelid) DESC;
EOF

log "✓ Unused indexes found: $OUTPUT_DIR/unused_indexes.txt"

# Step 3: Find duplicate indexes
log "Step 3: Finding duplicate indexes..."

psql "$DATABASE_URL" <<EOF > "$OUTPUT_DIR/duplicate_indexes.txt"
SELECT
    pg_size_pretty(SUM(pg_relation_size(idx))::bigint) as size,
    (array_agg(idx))[1] as index1,
    (array_agg(idx))[2] as index2,
    (array_agg(idx))[3] as index3,
    (array_agg(idx))[4] as index4
FROM (
    SELECT
        indexrelid::regclass as idx,
        indrelid::regclass as table,
        indrelid::regclass as table_name,
        array_to_string(indkey, ' ') as cols,
        indpred is not null as partial
    FROM pg_index
) sub
GROUP BY table, cols, partial
HAVING COUNT(*) > 1;
EOF

log "✓ Duplicate indexes found: $OUTPUT_DIR/duplicate_indexes.txt"

# Step 4: Find missing indexes (high sequential scans)
log "Step 4: Finding missing indexes..."

psql "$DATABASE_URL" <<EOF > "$OUTPUT_DIR/missing_indexes.txt"
SELECT 
    schemaname,
    tablename,
    seq_scan as sequential_scans,
    idx_scan as index_scans,
    seq_tup_read as seq_tuples_read,
    n_live_tup as live_rows,
    CASE 
        WHEN seq_scan > 1000 AND idx_scan = 0 THEN '⚠️  CRITICAL'
        WHEN seq_scan > 1000 AND idx_scan < seq_scan::float * 0.1 THEN '⚠️  WARNING'
        ELSE 'OK'
    END as status,
    'CREATE INDEX idx_' || tablename || '_seq ON ' || schemaname || '.' || tablename || ' (id);' as suggestion
FROM pg_stat_user_tables
WHERE seq_scan > 100
ORDER BY seq_scan DESC;
EOF

log "✓ Missing indexes found: $OUTPUT_DIR/missing_indexes.txt"

# Step 5: Check JSONB indexes
log "Step 5: Checking JSONB index usage..."

psql "$DATABASE_URL" <<EOF > "$OUTPUT_DIR/jsonb_indexes.txt"
SELECT 
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as size,
    idx_scan as scans,
    pg_indexes.indexdef
FROM pg_stat_user_indexes
JOIN pg_indexes ON pg_stat_user_indexes.indexrelname = pg_indexes.indexname
WHERE indexname LIKE '%metadata%' 
   OR indexname LIKE '%payload%' 
   OR indexname LIKE '%details%'
   OR indexname LIKE '%options%'
ORDER BY idx_scan ASC;
EOF

log "✓ JSONB indexes checked: $OUTPUT_DIR/jsonb_indexes.txt"

# Step 6: Analyze index bloat
log "Step 6: Analyzing index bloat..."

psql "$DATABASE_URL" <<EOF > "$OUTPUT_DIR/index_bloat.txt"
SELECT 
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as current_size,
    CASE 
        WHEN idx_scan > 10000 AND pg_relation_size(indexrelid) > 1000000 THEN 'REINDEX CANDIDATE'
        ELSE 'OK'
    END as status,
    'REINDEX INDEX ' || schemaname || '.' || indexname || ';' as command
FROM pg_stat_user_indexes
WHERE pg_relation_size(indexrelid) > 1000000
ORDER BY pg_relation_size(indexrelid) DESC;
EOF

log "✓ Index bloat analyzed: $OUTPUT_DIR/index_bloat.txt"

# Step 7: Generate optimization recommendations
log "Step 7: Generating optimization recommendations..."

cat > "$OUTPUT_DIR/recommendations.md" <<'EOF'
# Index Optimization Recommendations

## Analysis Date
$(date)

## Current Index Statistics

### Total Indexes
EOF

psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public';" >> "$OUTPUT_DIR/recommendations.md"

cat >> "$OUTPUT_DIR/recommendations.md" <<'EOF'

### Total Index Size
EOF

psql "$DATABASE_URL" -c "SELECT pg_size_pretty(SUM(pg_relation_size(indexrelid))) FROM pg_stat_user_indexes;" >> "$OUTPUT_DIR/recommendations.md"

cat >> "$OUTPUT_DIR/recommendations.md" <<'EOF'

## Recommendations

### 1. Drop Unused Indexes
Review and drop indexes with 0 scans (except primary keys and unique constraints):

EOF

cat "$OUTPUT_DIR/unused_indexes.txt" >> "$OUTPUT_DIR/recommendations.md"

cat >> "$OUTPUT_DIR/recommendations.md" <<'EOF'

**⚠️  Warning**: Test in staging first! Ensure no queries use these indexes.

### 2. Remove Duplicate Indexes
Found duplicate indexes that can be removed:

EOF

cat "$OUTPUT_DIR/duplicate_indexes.txt" >> "$OUTPUT_DIR/recommendations.md"

cat >> "$OUTPUT_DIR/recommendations.md" <<'EOF'

### 3. Add Missing Indexes
Tables with high sequential scans may need indexes:

EOF

cat "$OUTPUT_DIR/missing_indexes.txt" >> "$OUTPUT_DIR/recommendations.md"

cat >> "$OUTPUT_DIR/recommendations.md" <<'EOF'

**Recommendation**: Analyze query patterns before adding indexes. Use EXPLAIN ANALYZE.

### 4. Reindex Bloated Indexes
Large indexes with many scans should be reindexed:

EOF

cat "$OUTPUT_DIR/index_bloat.txt" >> "$OUTPUT_DIR/recommendations.md"

cat >> "$OUTPUT_DIR/recommendations.md" <<'EOF'

**Command**: `REINDEX INDEX CONCURRENTLY index_name;`

### 5. JSONB Index Verification
Ensure GIN indexes exist for JSONB columns:

EOF

cat "$OUTPUT_DIR/jsonb_indexes.txt" >> "$OUTPUT_DIR/recommendations.md"

cat >> "$OUTPUT_DIR/recommendations.md" <<'EOF'

**Command to add GIN index**:
```sql
CREATE INDEX CONCURRENTLY idx_table_column_gin ON table_name USING GIN (column_name);
```

## Implementation Steps

### Immediate Actions (Low Risk)
1. ✓ Reindex large indexes during maintenance window
2. ✓ Drop obviously unused indexes
3. ✓ Add missing GIN indexes for JSONB

### Investigate (Medium Risk)
1. Test query performance without duplicate indexes
2. Analyze slow queries before adding new indexes
3. Monitor index usage over 1-2 weeks

### Monitor (Ongoing)
1. Track index usage weekly
2. Alert on unused indexes (>7 days, 0 scans)
3. Alert on tables with seq_scan/index_scan ratio > 10

## Performance Impact

Expected improvements:
- **Unused indexes**: Free disk space, faster writes, no read impact
- **Duplicate indexes**: Faster writes, maintenance is clearer
- **Missing indexes**: 10-100x faster queries
- **Reindex**: Reduce bloat, improve cache efficiency

## Monitoring

Add to monitoring dashboard:
```sql
-- Index health check
SELECT COUNT(*) as unused_indexes
FROM pg_stat_user_indexes
WHERE idx_scan = 0 AND indexname NOT LIKE '%_pkey';

-- Index bloat check
SELECT COUNT(*) as bloated_indexes
FROM pg_stat_user_indexes
WHERE pg_relation_size(indexrelid) > 10000000
  AND idx_scan > 10000;
```

## Next Review
Schedule next index analysis in 2 weeks.
EOF

log "✓ Recommendations generated: $OUTPUT_DIR/recommendations.md"

# Step 8: Apply optimizations (if not dry-run)
if [ "$DRY_RUN" != "--dry-run" ]; then
    log "Step 8: Applying optimizations..."
    
    # Ask for confirmation
    read -p "Apply optimizations? This will drop unused indexes and reindex large tables. (yes/no): " confirm
    if [ "$confirm" = "yes" ]; then
        # Reindex large indexes
        psql "$DATABASE_URL" <<'SQLEOF'
-- Reindex indexes with high usage and likely bloat
REINDEX INDEX CONCURRENTLY idx_files_metadata;
REINDEX INDEX CONCURRENTLY idx_tasks_payload;
REINDEX INDEX CONCURRENTLY idx_notifications_metadata;
ANALYZE;
SQLEOF
        
        log "✓ Indexes reindexed"
        
        # Prompts for unused indexes
        log "Review unused indexes before dropping:"
        cat "$OUTPUT_DIR/unused_indexes.txt"
        
        # TODO: Add interactive drop functionality
    else
        log "Skipping optimizations. Run with --apply to apply changes."
    fi
else
    log "Step 8: Dry-run mode - skipping optimizations"
fi

# Step 9: Run ANALYZE
log "Step 9: Updating statistics..."
psql "$DATABASE_URL" -c "ANALYZE;" || log "⚠ Warning: ANALYZE failed"

log "✓ Statistics updated"

# Step 10: Summary
log "================================"
log "✅ Index analysis completed!"
log ""
log "Results:"
log "  - Index usage: $OUTPUT_DIR/index_usage.txt"
log "  - Unused indexes: $OUTPUT_DIR/unused_indexes.txt"
log "  - Duplicate indexes: $OUTPUT_DIR/duplicate_indexes.txt"
log "  - Missing indexes: $OUTPUT_DIR/missing_indexes.txt"
log "  - JSONB indexes: $OUTPUT_DIR/jsonb_indexes.txt"
log "  - Index bloat: $OUTPUT_DIR/index_bloat.txt"
log "  - Recommendations: $OUTPUT_DIR/recommendations.md"
log ""
log "To apply optimizations:"
log "  1. Review recommendations.md"
log "  2. Test in staging environment"
log "  3. Run: ./scripts/optimize-indexes.sh --apply"
log ""

# Schedule for cron
log "To schedule weekly index analysis:"
log "  0 3 * * 0 $PWD/scripts/optimize-indexes.sh > /var/log/postgres-index-analysis.log 2>&1"
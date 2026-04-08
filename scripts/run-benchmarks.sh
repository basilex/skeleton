#!/bin/bash
# Run PostgreSQL performance benchmarks
# Usage: ./scripts/run-benchmarks.sh [benchmark_name]

set -e

echo "📊 PostgreSQL Performance Benchmarks"
echo "======================================"

# Configuration
BENCHMARK_DB="${BENCHMARK_DATABASE_URL:-postgres://skeleton:password@localhost:5432/skeleton_benchmark}"
RESULTS_DIR="./benchmark_results/$(date +%Y%m%d_%H%M%S)"
BASELINE_FILE="./benchmark_results/baseline.txt"

# Create results directory
mkdir -p "$RESULTS_DIR"

# Log function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Check if benchmark binary exists
if [ ! -f "./internal/benchmark/postgres_queries_benchmark_test.go" ]; then
    log "❌ Benchmark file not found"
    exit 1
fi

# Step 1: Setup test database
log "Step 1: Setting up test database..."

# Drop and recreate test database
psql postgres://skeleton:password@localhost:5432/postgres -c "DROP DATABASE IF EXISTS skeleton_benchmark;" 2>/dev/null || true
psql postgres://skeleton:password@localhost:5432/postgres -c "CREATE DATABASE skeleton_benchmark;" || {
    log "⚠ Warning: Could not create test database. Using existing."
}

# Run migrations
log "Running migrations on test database..."
psql "$BENCHMARK_DB" -f migrations/001_initial_schema.sql > /dev/null 2>&1 || log "⚠ Warning: Schema migration partially applied"

log "✓ Test database ready"

# Step 2: Seed test data
log "Step 2: Seeding test data..."

# Generate seed data using psql
psql "$BENCHMARK_DB" <<EOF
-- Seed users (1000)
INSERT INTO users (id, email, password_hash, is_active, created_at, updated_at)
SELECT 
    uuid_generate_v7(),
    'user' || i || '@example.com',
    '\$2a\$10\$hash' || i,
    true,
    NOW() - (random() * '365 days'::interval),
    NOW()
FROM generate_series(1, 1000) AS i;

-- Seed files (10000)
INSERT INTO files (id, owner_id, filename, stored_name, mime_type, size, path, storage_provider, checksum, metadata, access_level, uploaded_at, created_at, updated_at)
SELECT 
    uuid_generate_v7(),
    (SELECT id FROM users ORDER BY random() LIMIT 1),
    'file' || i || '.jpg',
    'stored' || i || '.jpg',
    CASE WHEN random() < 0.5 THEN 'image/jpeg' ELSE 'image/png' END,
    (random() * 10000000)::int,
    '/files/file' || i || '.jpg',
    'local',
    'sha256:abc' || i,
    jsonb_build_object('width', (random() * 1000)::int, 'height', (random() * 1000)::int),
    CASE WHEN random() < 0.7 THEN 'private' ELSE 'public' END,
    NOW() - (random() * '365 days'::interval),
   NOW() - (random() * '365 days'::interval),
    NOW() - (random() * '30 days'::interval)
FROM generate_series(1, 10000) AS i;

-- Seed tasks (50000)
INSERT INTO tasks (id, type, payload, status, priority, scheduled_at, attempts, max_attempts, created_at, updated_at)
SELECT 
    uuid_generate_v7(),
    CASE (i % 4) 
        WHEN 0 THEN 'send_email'
        WHEN 1 THEN 'process_file'
        WHEN 2 THEN 'generate_report'
        ELSE 'cleanup_data'
    END,
    jsonb_build_object('key', 'value' || i),
    CASE (random() < 0.1)
        WHEN true THEN 'pending'
        WHEN random() < 0.3 THEN 'running'
        ELSE 'completed'
    END,
    (random() * 10)::int,
    NOW() - (random() * '7 days'::interval),
    (random() * 3)::int,
    3,
    NOW() - (random() * '30 days'::interval),
    NOW() - (random() * '30 days'::interval)
FROM generate_series(1, 50000) AS i;

-- Seed notifications (20000)
INSERT INTO notifications (id, user_id, channel, subject, content, status, priority, created_at, updated_at)
SELECT 
    uuid_generate_v7(),
    (SELECT id FROM users ORDER BY random() LIMIT 1),
    CASE (i % 4)
        WHEN 0 THEN 'email'
        WHEN 1 THEN 'sms'
        WHEN 2 THEN 'push'
        ELSE 'in_app'
    END,
    'Subject ' || i,
    'Content for notification ' || i,
    CASE 
        WHEN random() < 0.3 THEN 'pending'
        WHEN random() < 0.6 THEN 'sent'
        ELSE 'delivered'
    END,
    CASE WHEN random() < 0.1 THEN 'high' ELSE 'normal' END,
    NOW() - (random() * '30 days'::interval),
    NOW() - (random() * '30 days'::interval)
FROM generate_series(1, 20000) AS i;

-- Analyze tables
ANALYZE;
EOF

log "✓ Test data seeded"

# Step 3: Run benchmarks
log "Step 3: Running benchmarks..."

BENCHMARK_PATTERN="${1:-.*}"

# Run benchmarks and capture output
go test -bench="$BENCHMARK_PATTERN" -benchmem -benchtime=5s \
    -run=^$ ./internal/benchmark \
    | tee "$RESULTS_DIR/benchmark_results.txt"

BENCHMARK_EXIT_CODE=${PIPESTATUS[0]}

if [ $BENCHMARK_EXIT_CODE -ne 0 ]; then
    log "❌ Benchmarks failed with exit code $BENCHMARK_EXIT_CODE"
    exit 1
fi

log "✓ Benchmarks completed"

# Step 4: Parse results
log "Step 4: Parsing results..."

# Extract key metrics
grep -E "Benchmark" "$RESULTS_DIR/benchmark_results.txt" | while read -r line; do
    log "  $line"
done

# Step 5: Compare with baseline
if [ -f "$BASELINE_FILE" ]; then
    log "Step 5: Comparing with baseline..."
    
    # Compare results
    if command -v benchstat &> /dev/null; then
        benchstat "$BASELINE_FILE" "$RESULTS_DIR/benchmark_results.txt" > "$RESULTS_DIR/comparison.txt"
        log "Comparison saved to: $RESULTS_DIR/comparison.txt"
        
        # Show comparison
        cat "$RESULTS_DIR/comparison.txt"
    else
        log "⚠ benchstat not installed. Install with: go install golang.org/x/perf/cmd/benchstat@latest"
    fi
else
    log "Step 5: No baseline found. Creating baseline..."
    cp "$RESULTS_DIR/benchmark_results.txt" "$BASELINE_FILE"
    log "✓ Baseline created at: $BASELINE_FILE"
fi

# Step 6: Database metrics
log "Step 6: Collecting database metrics..."

psql "$BENCHMARK_DB" <<EOF > "$RESULTS_DIR/database_stats.txt"
-- Index usage
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Cache hit ratio
SELECT 
    'index hit ratio' as metric,
    sum(idx_blks_hit)::float / (sum(idx_blks_hit) + sum(idx_blks_read)) * 100 as value
FROM pg_statio_user_indexes
UNION ALL
SELECT 
    'table hit ratio' as metric,
    sum(heap_blks_hit)::float / (sum(heap_blks_hit) + sum(heap_blks_read)) * 100 as value
FROM pg_statio_user_tables;

-- Table sizes
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
    n_live_tup,
    n_dead_tup
FROM pg_stat_user_tables
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Slow queries simulation
SELECT 'File by ID' as query, AVG(extract('milliseconds' from now())) as avg_ms
FROM (SELECT 1 FROM files WHERE id = uuid_generate_v7() LIMIT 100) t
UNION ALL
SELECT 'JSONB search' as query, AVG(extract('milliseconds' from now())) as avg_ms
FROM (SELECT 1 FROM files WHERE metadata @> '{"width":500}' LIMIT 100) t;
EOF

log "✓ Database metrics saved to: $RESULTS_DIR/database_stats.txt"

# Step 7: Generate report
log "Step 7: Generating benchmark report..."

cat > "$RESULTS_DIR/report.md" <<EOF
# PostgreSQL Benchmark Report

Date: $(date)
Database: $BENCHMARK_DB
Data Volume: 
- Users: 1,000
- Files: 10,000
- Tasks: 50,000
- Notifications: 20,000

## Benchmark Results

\`\`\`
$(cat "$RESULTS_DIR/benchmark_results.txt")
\`\`\`

## Database Statistics

\`\`\`
$(cat "$RESULTS_DIR/database_stats.txt")
\`\`\`

## Performance Targets

| Benchmark | Target | Actual | Status |
|-----------|--------|--------|--------|
| GetFileByID | < 1ms | TBD | - |
| ListFilesByOwner | < 5ms | TBD | - |
| SearchFilesWithJSONB | < 10ms | TBD | - |
| TaskQueue | < 2ms | TBD | - |
| NotificationBatch | < 50ms | TBD | - |
| AuditLogInsert | < 1ms | TBD | - |

## Recommendations

1. Review slow queries (>10ms)
2. Check index usage for JSONB queries
3. Verify materialized views are refreshed
4. Monitor cache hit ratios (>99% expected)

EOF

log "✓ Report generated: $RESULTS_DIR/report.md"

# Step 8: Cleanup
log "Step 8: Cleanup..."

# Optionally drop test database
read -p "Drop test database? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    psql postgres://skeleton:password@localhost:5432/postgres -c "DROP DATABASE skeleton_benchmark;" || true
    log "✓ Test database dropped"
fi

log "======================================="
log "✅ Benchmarks completed successfully!"
log "Results: $RESULTS_DIR"
log "Report: $RESULTS_DIR/report.md"
log "======================================="

# Step 9: Set up continuous benchmarking
log ""
log "To set up continuous benchmarking:"
log "  1. Add to CI/CD pipeline: .github/workflows/benchmark.yml"
log "  2. Schedule nightly benchmarks"
log "  3. Alert on performance regression (>20% slower)"
log ""
log "Example CI configuration:"
cat <<'YAML'
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
          POSTGRES_USER: skeleton
          POSTGRES_PASSWORD: password
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      - run: ./scripts/run-benchmarks.sh
      - uses: actions/upload-artifact@v3
        with:
          name: benchmark-results
          path: benchmark_results/
YAML
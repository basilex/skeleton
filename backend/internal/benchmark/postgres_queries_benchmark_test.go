package benchmark

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BenchmarkFileRepositoryQueries benchmarks file repository queries
// Run with: go test -bench=. -benchmem ./internal/benchmark

// BenchmarkGetFileByID benchmarks retrieving a file by ID
func BenchmarkGetFileByID(b *testing.B) {
	pool := setupBenchmarkDB(b)
	defer pool.Close()

	ctx := context.Background()
	fileID := createTestFile(b, pool)
	defer cleanupTestFile(b, pool, fileID)

	query := `SELECT id, owner_id, filename, stored_name, mime_type, size, path, 
              storage_provider, checksum, metadata, access_level, uploaded_at, 
              expires_at, processed_at, created_at, updated_at
              FROM files WHERE id = $1`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var id, ownerID, filename, storedName, mimeType, path, provider, checksum, accessLevel string
		var size int64
		var uploadedAt, createdAt, updatedAt time.Time
		var expiresAt, processedAt *time.Time
		var metadata []byte

		err := pool.QueryRow(ctx, query, fileID).Scan(
			&id, &ownerID, &filename, &storedName, &mimeType, &size, &path,
			&provider, &checksum, &metadata, &accessLevel,
			&uploadedAt, &expiresAt, &processedAt, &createdAt, &updatedAt,
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListFilesByOwner benchmarks listing files by owner with JSONB metadata query
func BenchmarkListFilesByOwner(b *testing.B) {
	pool := setupBenchmarkDB(b)
	defer pool.Close()

	ctx := context.Background()
	ownerID := createTestUser(b, pool)
	for i := 0; i < 100; i++ {
		createTestFileForOwner(b, pool, ownerID)
	}
	defer cleanupTestUser(b, pool, ownerID)

	query := `SELECT id, owner_id, filename, stored_name, mime_type, size, path,
              storage_provider, checksum, metadata, access_level, uploaded_at
              FROM files
              WHERE owner_id = $1
              ORDER BY uploaded_at DESC
              LIMIT 20`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := pool.Query(ctx, query, ownerID)
		if err != nil {
			b.Fatal(err)
		}
		rows.Close()
	}
}

// BenchmarkSearchFilesWithJSONB benchmarks searching files using JSONB queries
func BenchmarkSearchFilesWithJSONB(b *testing.B) {
	pool := setupBenchmarkDB(b)
	defer pool.Close()

	ctx := context.Background()
	// Create files with metadata
	for i := 0; i < 100; i++ {
		createTestFileWithMetadata(b, pool, fmt.Sprintf(`{"width":%d,"height":%d}`, i%1000+1, i%1000+1))
	}

	// Query with JSONB containment operator
	query := `SELECT id, filename FROM files 
              WHERE metadata @> '{"width": 500}'::jsonb
              LIMIT 20`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows, err := pool.Query(ctx, query)
		if err != nil {
			b.Fatal(err)
		}
		rows.Close()
	}
}

// BenchmarkTaskQueue benchmarks task queue queries
func BenchmarkTaskQueue(b *testing.B) {
	pool := setupBenchmarkDB(b)
	defer pool.Close()

	ctx := context.Background()
	// Create pending tasks
	for i := 0; i < 1000; i++ {
		createTestTask(b, pool, "pending")
	}
	defer cleanupTasks(b, pool)

	// Simulate worker fetching next task
	query := `UPDATE tasks
              SET status = 'running', started_at = NOW(), updated_at = NOW()
              WHERE id = (
                  SELECT id FROM tasks
                  WHERE status = 'pending'
                  AND scheduled_at <= NOW()
                  ORDER BY priority DESC, created_at ASC
                  LIMIT 1
                  FOR UPDATE SKIP LOCKED
              )
              RETURNING id`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var id string
		err := pool.QueryRow(ctx, query).Scan(&id)
		if err != nil {
			// No more tasks
			break
		}
	}
}

// BenchmarkNotificationBatch benchmarks creating notifications in batch
func BenchmarkNotificationBatch(b *testing.B) {
	pool := setupBenchmarkDB(b)
	defer pool.Close()

	ctx := context.Background()
	userIDs := make([]string, 100)
	for i := 0; i < 100; i++ {
		userIDs[i] = createTestUser(b, pool)
	}
	defer func() {
		for _, id := range userIDs {
			cleanupTestUser(b, pool, id)
		}
	}()

	query := `INSERT INTO notifications 
              (id, user_id, channel, subject, content, status, priority, created_at, updated_at)
              VALUES (uuid_generate_v7(), $1, $2, $3, $4, 'pending', 'normal', NOW(), NOW())`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch := make([]string, 100)
		for j := 0; j < 100; j++ {
			batch[j] = fmt.Sprintf("%s,%s,%s,%s",
				userIDs[j%len(userIDs)],
				"email",
				"Test Subject",
				"Test Content",
			)
		}
		// Batch insert
		for _, params := range batch {
			_, err := pool.Exec(ctx, query, params)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkAuditLogInsert benchmarks audit log insertions
func BenchmarkAuditLogInsert(b *testing.B) {
	pool := setupBenchmarkDB(b)
	defer pool.Close()

	ctx := context.Background()
	userID := createTestUser(b, pool)
	defer cleanupTestUser(b, pool, userID)

	query := `INSERT INTO audit_records 
              (id, user_id, actor_type, action, resource, resource_id, 
               details, ip_address, user_agent, created_at)
              VALUES (uuid_generate_v7(), $1, $2, $3, $4, $5, $6, $7, $8, NOW())`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pool.Exec(ctx, query,
			userID,
			"user",
			"login",
			"session",
			fmt.Sprintf("session-%d", i),
			`{"browser":"chrome","os":"linux"}`,
			"192.168.1.1",
			"Mozilla/5.0",
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper functions

func setupBenchmarkDB(b *testing.B) *pgxpool.Pool {
	// This would connect to a test database
	// In production, use environment variables or testcontainers
	connStr := "postgres://user:pass@localhost:5432/skeleton_test?sslmode=disable"

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		b.Skipf("Failed to connect to database: %v", err)
	}
	return pool
}

func createTestUser(b *testing.B, pool *pgxpool.Pool) string {
	ctx := context.Background()
	user, err := domain.NewUser(domain.Email("test@example.com"), "hash")
	if err != nil {
		b.Fatal(err)
	}

	var id string
	err = pool.QueryRow(ctx,
		`INSERT INTO users (id, email, password_hash, is_active, created_at, updated_at) 
         VALUES (uuid_generate_v7(), $1, $2, $3, NOW(), NOW()) 
         RETURNING id`,
		user.Email().String(), "hashed_password", true,
	).Scan(&id)
	if err != nil {
		b.Fatal(err)
	}
	return id
}

func createTestFile(b *testing.B, pool *pgxpool.Pool) string {
	// Create test file
	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO files (id, filename, stored_name, mime_type, size, path, 
         storage_provider, checksum, metadata, access_level, uploaded_at, created_at, updated_at)
         VALUES (uuid_generate_v7(), 'test.jpg', 'abc123.jpg', 'image/jpeg', 1024, '/files/test.jpg',
         'local', 'sha256:abc', '{}', 'private', NOW(), NOW(), NOW())
         RETURNING id`,
	).Scan(&id)
	if err != nil {
		b.Fatal(err)
	}
	return id
}

func createTestFileForOwner(b *testing.B, pool *pgxpool.Pool, ownerID string) {
	_, err := pool.Exec(context.Background(),
		`INSERT INTO files (id, owner_id, filename, stored_name, mime_type, size, path,
         storage_provider, checksum, metadata, access_level, uploaded_at, created_at, updated_at)
         VALUES (uuid_generate_v7(), $1, 'test.jpg', 'abc.jpg', 'image/jpeg', 1024, '/files/test.jpg',
         'local', 'sha256:abc', '{}', 'private', NOW(), NOW(), NOW())`,
		ownerID,
	)
	if err != nil {
		b.Fatal(err)
	}
}

func createTestFileWithMetadata(b *testing.B, pool *pgxpool.Pool, metadata string) {
	_, err := pool.Exec(context.Background(),
		`INSERT INTO files (id, filename, stored_name, mime_type, size, path,
         storage_provider, checksum, metadata, access_level, uploaded_at, created_at, updated_at)
         VALUES (uuid_generate_v7(), 'test.jpg', 'abc.jpg', 'image/jpeg', 1024, '/files/test.jpg',
         'local', 'sha256:abc', $1, 'private', NOW(), NOW(), NOW())`,
		metadata,
	)
	if err != nil {
		b.Fatal(err)
	}
}

func createTestTask(b *testing.B, pool *pgxpool.Pool, status string) {
	_, err := pool.Exec(context.Background(),
		`INSERT INTO tasks (id, type, payload, status, attempts, max_attempts, scheduled_at, created_at, updated_at)
         VALUES (uuid_generate_v7(), 'send_email', '{}', $1, 0, 3, NOW(), NOW(), NOW())`,
		status,
	)
	if err != nil {
		b.Fatal(err)
	}
}

func cleanupTestFile(b *testing.B, pool *pgxpool.Pool, id string) {
	_, err := pool.Exec(context.Background(), "DELETE FROM files WHERE id = $1", id)
	if err != nil {
		b.Fatal(err)
	}
}

func cleanupTestUser(b *testing.B, pool *pgxpool.Pool, id string) {
	_, err := pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		b.Fatal(err)
	}
}

func cleanupTasks(b *testing.B, pool *pgxpool.Pool) {
	_, err := pool.Exec(context.Background(), "DELETE FROM tasks WHERE 1=1")
	if err != nil {
		b.Fatal(err)
	}
}

package health

import (
	"context"
	"database/sql"
	"time"
)

// Status represents the health status of a component.
type Status string

const (
	StatusOK            Status = "ok"             // Component is healthy
	StatusError         Status = "error"          // Component is unhealthy
	StatusNotConfigured Status = "not_configured" // Component is not configured
	StatusDisabled      Status = "disabled"       // Component is explicitly disabled
)

// ComponentHealth represents the health status of a system component.
type ComponentHealth struct {
	Status    Status `json:"status"`
	Type      string `json:"type"`
	Connected bool   `json:"connected,omitempty"`
	Latency   string `json:"latency,omitempty"`
	Error     string `json:"error,omitempty"`
	Version   string `json:"version,omitempty"`
}

// SystemHealth represents the overall system health.
type SystemHealth struct {
	Database  ComponentHealth `json:"database"`
	Cache     ComponentHealth `json:"cache"`
	RateLimit ComponentHealth `json:"rate_limit"`
}

// Checker provides health checking functionality.
type Checker struct {
	db     *sql.DB
	dbType string
}

// NewChecker creates a new health checker.
func NewChecker(db *sql.DB, dbType string) *Checker {
	return &Checker{
		db:     db,
		dbType: dbType,
	}
}

// CheckDatabase checks the database connection health.
func (c *Checker) CheckDatabase(ctx context.Context) ComponentHealth {
	health := ComponentHealth{
		Type: c.dbType,
	}

	start := time.Now()
	err := c.db.PingContext(ctx)
	latency := time.Since(start)

	if err != nil {
		health.Status = StatusError
		health.Connected = false
		health.Error = err.Error()
		return health
	}

	health.Status = StatusOK
	health.Connected = true
	health.Latency = latency.String()

	// Get database version if available
	var version string
	row := c.db.QueryRowContext(ctx, "SELECT sqlite_version()")
	if err := row.Scan(&version); err == nil {
		health.Version = version
	}

	return health
}

// CheckRedis checks Redis connection health.
// Returns StatusNotConfigured if Redis is not enabled in current config.
func (c *Checker) CheckRedis(ctx context.Context, redisURL string) ComponentHealth {
	health := ComponentHealth{
		Type: "redis",
	}

	// If Redis URL is empty or not configured, return not configured
	if redisURL == "" || redisURL == "redis://localhost:6379/0" && c.dbType == "sqlite" {
		// Development mode - Redis not required
		health.Status = StatusNotConfigured
		health.Connected = false
		health.Error = "not required for development mode (in-memory event bus)"
		return health
	}

	// TODO: Add actual Redis connection check when Redis is integrated
	// For now, return not configured
	health.Status = StatusNotConfigured
	health.Connected = false
	health.Error = "Redis integration pending"

	return health
}

// CheckCache checks the cache health.
func (c *Checker) CheckCache(ctx context.Context, cacheType string) ComponentHealth {
	health := ComponentHealth{
		Type: cacheType,
	}

	switch cacheType {
	case "memory":
		health.Status = StatusOK
		health.Connected = true
		health.Error = "in-memory cache (no external connection required)"
	case "redis":
		// TODO: Add actual Redis cache health check
		health.Status = StatusNotConfigured
		health.Connected = false
		health.Error = "Redis cache not yet implemented"
	default:
		health.Status = StatusError
		health.Error = "unknown cache type"
	}

	return health
}

// CheckRateLimiter checks the rate limiter health.
func (c *Checker) CheckRateLimiter(ctx context.Context, rateLimitType string, enabled bool) ComponentHealth {
	health := ComponentHealth{
		Type: rateLimitType,
	}

	if !enabled {
		health.Status = StatusDisabled
		health.Connected = false
		health.Error = "rate limiting is disabled"
		return health
	}

	switch rateLimitType {
	case "token_bucket":
		health.Status = StatusOK
		health.Connected = true
		health.Error = "in-memory rate limiter (no external connection required)"
	case "sliding_window":
		// TODO: Add actual Redis rate limiter health check
		health.Status = StatusNotConfigured
		health.Connected = false
		health.Error = "Redis rate limiter not yet implemented"
	default:
		health.Status = StatusError
		health.Error = "unknown rate limiter type"
	}

	return health
}

// FullCheck performs a complete system health check.
func (c *Checker) FullCheck(ctx context.Context, cacheType, rateLimitType string, rateLimitEnabled bool, redisURL string) SystemHealth {
	return SystemHealth{
		Database:  c.CheckDatabase(ctx),
		Cache:     c.CheckCache(ctx, cacheType),
		RateLimit: c.CheckRateLimiter(ctx, rateLimitType, rateLimitEnabled),
	}
}

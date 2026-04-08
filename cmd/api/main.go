// Package main provides the HTTP API server for the Skeleton application.
// This package implements a RESTful API following Domain-Driven Design (DDD)
// and hexagonal architecture principles. It serves as the entry point for the
// application, handling dependency injection, routing configuration, and server lifecycle.
//
// The API provides the following capabilities:
//   - Authentication and authorization (session-based with JWT tokens)
//   - User management (CRUD operations with role-based access control)
//   - Role management and permission handling
//   - Audit logging for tracking user actions
//   - Notification system with background worker processing
//   - System status and health checks
//
// Architecture Overview:
//
// The application follows a hexagonal (ports and adapters) architecture with clear separation of concerns:
//   - Domain layer: Core business logic and entities
//   - Application layer: Use cases, commands, and queries
//   - Infrastructure layer: Persistence, external services, and adapters
//   - Ports layer: HTTP handlers and DTOs
//
// The application uses dependency injection (wire.go) to construct the dependency graph
// and Gin as the HTTP router with middleware for cross-cutting concerns.
//
// Swagger Documentation:
//
// @title Skeleton API
// @version 1.0
// @description Go DDD Hexagonal architecture skeleton project API
// @termsOfService http://swagger.io/terms/
//
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
//
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host localhost:8080
// @BasePath /
//
// @securityDefinitions.apikey SessionAuth
// @in cookie
// @name session
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token (format: "Bearer {token}")
//
// @tag.name auth
// @tag.description Authentication operations
// @tag.name users
// @tag.description User management operations
// @tag.name roles
// @tag.description Role management operations
// @tag.name audit
// @tag.description Audit log operations
// @tag.name notifications
// @tag.description Notification management operations
// @tag.name tasks
// @tag.description Background tasks and scheduled jobs
// @tag.name status
// @tag.description System status and health checks
package main

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/basilex/skeleton/pkg/config"
	"github.com/basilex/skeleton/pkg/database"
	"github.com/basilex/skeleton/pkg/httpserver"
	redisclient "github.com/basilex/skeleton/pkg/redis"
	"github.com/redis/go-redis/v9"
)

// Build information populated at compile time via ldflags.
// These variables provide version tracking and build metadata for
// observability and debugging purposes.
// They are set via ldflags during build, but have sensible defaults for development.
var (
	// version represents the application version (e.g., "1.0.0").
	version = "dev"
	// commit represents the git commit hash at build time.
	commit = "none"
	// buildTime represents the timestamp when the binary was built.
	buildTime = "unknown"
)

func parseBuildTime() time.Time {
	if buildTime == "unknown" || buildTime == "" {
		return time.Now()
	}
	t, err := time.Parse(time.RFC3339, buildTime)
	if err != nil {
		return time.Now()
	}
	return t
}

// main is the application entry point. It orchestrates the entire application
// lifecycle including configuration loading, database initialization, dependency
// wiring, and HTTP server management.
//
// The startup sequence is:
//  1. Load configuration from environment/file
//  2. Initialize structured logging
//  3. Connect to PostgreSQL database
//  4. Connect to Redis (optional)
//  5. Wire all dependencies (repositories, services, handlers)
//  6. Start background notification worker
//  7. Configure HTTP router with middleware and routes
//  8. Start HTTP server
//  9. Wait for shutdown signal
//  10. Gracefully shutdown server with timeout
//
// The function exits with code 1 if any critical initialization step fails.
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	setupLogger(cfg)

	// Connect to PostgreSQL
	pool, err := database.NewPostgresPool(database.PostgresConfig{
		URL:             cfg.Database.URL,
		MaxConns:        int32(cfg.Database.MaxOpenConns),
		MinConns:        int32(cfg.Database.MaxIdleConns),
		MaxConnLifetime: 5 * time.Minute,
		MaxConnIdleTime: 10 * time.Minute,
		HealthCheck:     1 * time.Minute,
	})
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Connect to Redis (optional, for cache/event bus)
	var redisClient *redis.Client
	if cfg.Redis.URL != "" {
		redisClient, err = redisclient.NewClient(redisclient.Config{URL: cfg.Redis.URL})
		if err != nil {
			slog.Error("failed to connect to redis", "error", err)
			os.Exit(1)
		}
		defer redisClient.Close()
	}

	di := wireDependencies(cfg, pool, redisClient, version, commit, buildTime, runtime.Version())

	// Start notification worker in background
	go func() {
		if err := di.NotificationWorker.Start(context.Background()); err != nil {
			slog.Error("notification worker error", "error", err)
		}
	}()
	slog.Info("notification worker started")

	router := setupRouter(cfg, di)

	srv := httpserver.New(router, cfg.App.Port)
	if err := srv.Start(); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}

	<-httpserver.WaitForShutdown(context.Background())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}

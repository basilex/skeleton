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
//   - Parties management (customers, suppliers, partners, employees)
//   - Contracts management
//   - Accounting (chart of accounts, transactions)
//   - Ordering (orders, quotes)
//   - Catalog management (items, categories)
//   - Invoicing (invoices, payments)
//   - Inventory management (warehouses, stock, movements, reservations)
//   - Document management (PDF generation, signatures)
//   - Files management (upload, download, processing)
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
// @title Skeleton Business Engine API
// @version 2.0
// @description Enterprise-grade RESTful API for business management including parties, contracts, accounting, ordering, catalog, invoicing, inventory, and documents.
// @description Cross-context integration via domain events enables seamless workflow automation.
// @termsOfService http://swagger.io/terms/
//
// @contact.name API Support
// @contact.url https://github.com/basilex/skeleton/issues
// @contact.email support@skeleton.local
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /
//
// @securityDefinitions.apikey SessionAuth
// @in cookie
// @name session
// @description Session-based authentication for web clients
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token authentication (format: "Bearer {token}")
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
		MaxConnLifetime: cfg.Database.ConnMaxLifetime,
		MaxConnIdleTime: cfg.Database.ConnMaxIdleTime,
		HealthCheck:     cfg.Database.HealthCheck,
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

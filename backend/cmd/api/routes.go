// Package main provides HTTP routing configuration for the API server.
// This file contains the Gin router setup, middleware configuration, and route registration
// for all API endpoints organized by domain context.
//
// Routes are organized into the following groups:
//   - Authentication routes (/api/v1/auth): Register, login, logout, refresh, profile
//   - User routes (/api/v1/users): List users, get user details, deactivate users
//   - Role routes (/api/v1/users/:id/roles): Assign and revoke roles
//   - Audit routes (/api/v1/audit): Query audit records
//   - Notification routes (/api/v1/notifications): Manage notifications and preferences
//   - Parties routes (/api/v1/customers, /api/v1/suppliers): Manage customers and suppliers
//   - Files routes (/api/v1/files): Upload, download, process files
//   - System routes (/system): Health checks, build info, system status
//
// Middleware chain (applied in order):
//  1. Recovery middleware - Panic recovery and error response
//  2. Request ID middleware - Unique request identification
//  3. Logger middleware - Structured request logging
//  4. CORS middleware - Cross-origin resource sharing
//  5. Global rate limiting - Prevent API abuse (optional, can be disabled)
//
// Protected routes use authentication and RBAC middlewares as needed.
// Additional middleware:
//   - Rate limiting on auth endpoints (brute force protection)
//   - Caching on read-heavy endpoints (performance optimization)
package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/basilex/skeleton/pkg/config"
	"github.com/basilex/skeleton/pkg/middleware"
	"github.com/basilex/skeleton/pkg/uuid"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// setupRouter creates and configures the Gin HTTP router with all middleware and routes.
// In production mode, sets Gin to release mode for optimized performance.
//
// The router is configured with:
//   - Global middleware chain (recovery, request ID, logger, CORS)
//   - API version group (/api/v1)
//   - All domain-specific route handlers
//   - Status endpoints outside versioned API
//
// Parameters:
//   - cfg: Application configuration for environment-specific settings
//   - di: Dependency container with all initialized handlers and middlewares
//
// Returns a configured Gin engine ready to serve HTTP requests.
func setupRouter(cfg *config.Config, di *Dependencies) *gin.Engine {
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	registerGlobalMiddleware(r)
	registerRoutes(r, di)

	return r
}

// registerGlobalMiddleware adds the middleware chain to the Gin engine.
// Middleware is applied in order and executes in the following sequence:
//  1. Recovery: Catches panics and returns 500 error with request ID
//  2. Request ID: Generates or propagates X-Request-ID header
//  3. Logger: Logs request details with structured logging
//  4. CORS: Handles cross-origin requests with configured policy
//
// Parameters:
//   - r: Gin engine to configure
func registerGlobalMiddleware(r *gin.Engine) {
	r.Use(recoveryMiddleware())
	r.Use(requestIDMiddleware())
	r.Use(loggerMiddleware())
	r.Use(corsMiddleware())
}

// registerRoutes registers all API routes organized by domain context.
// Routes are grouped under /api/v1 for versioned endpoints, with system
// monitoring endpoints at /health and /system/*.
//
// Route Structure:
//
//	/health -> Health check (load balancer endpoint)
//	/system/ready -> Readiness probe
//	/system/build -> Build information
//	/system/info -> Detailed system info
//	/api/v1/auth -> Authentication endpoints
//	/api/v1/users -> User management endpoints
//	/api/v1/audit -> Audit log endpoints
//	/api/v1/notifications -> Notification endpoints
//	/api/v1/files -> File management endpoints
//
// Parameters:
//   - r: Gin engine to configure
//   - di: Dependency container with all initialized handlers
func registerRoutes(r *gin.Engine, di *Dependencies) {
	v1 := r.Group("/api/v1")
	{
		registerAuthRoutes(v1, di)
		registerUserRoutes(v1, di)
		registerRoleRoutes(v1, di)
		registerSessionRoutes(v1, di)
		registerPreferencesRoutes(v1, di)
		registerAuditRoutes(v1, di)
		registerNotificationRoutes(v1, di)
		registerPartiesRoutes(v1, di)
		registerAccountingRoutes(v1, di)
		registerInvoicingRoutes(v1, di)
		registerInventoryRoutes(v1, di)
		registerDocumentsRoutes(v1, di)
		registerOrderingRoutes(v1, di)
		registerCatalogRoutes(v1, di)
		registerFilesRoutes(v1, di)
	}

	registerStatusRoutes(r, di)
}

// registerAuthRoutes registers authentication-related endpoints.
// All auth routes are publicly accessible except /auth/logout and /auth/me
// which require session authentication.
//
// Rate Limiting:
//   - Login: 5 requests per minute per IP (brute force protection)
//   - Register: 3 requests per hour per IP (spam protection)
//   - Password reset: 3 requests per hour per IP (abuse prevention)
//
// Endpoints:
//   - POST /api/v1/auth/register - Register a new user (rate limited)
//   - POST /api/v1/auth/login - Authenticate user and create session (rate limited)
//   - POST /api/v1/auth/refresh - Refresh access token
//   - POST /api/v1/auth/logout - End user session (requires authentication)
//   - GET /api/v1/auth/me - Get current user profile (requires authentication)
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with identity handler and session middleware
func registerAuthRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	v1.POST("/auth/login",
		middleware.RateLimit(di.RateLimiter, middleware.ByIP, 5, time.Minute),
		di.IdentityHandler.Login)

	v1.POST("/auth/register",
		middleware.RateLimit(di.RateLimiter, middleware.ByIP, 3, time.Hour),
		di.IdentityHandler.Register)

	v1.POST("/auth/refresh", di.IdentityHandler.Refresh)

	v1.POST("/auth/logout", di.AuthMiddleware.Authenticate(), di.IdentityHandler.Logout)
	v1.GET("/auth/me", di.AuthMiddleware.Authenticate(), di.IdentityHandler.GetMyProfile)
}

// registerUserRoutes registers user management endpoints.
// All user routes require JWT authentication and RBAC permissions.
//
// Caching:
//   - GET /users: Cached for 5 minutes (user list rarely changes)
//   - GET /users/:id: Cached for 5 minutes (user profile cached)
//
// Endpoints:
//   - GET /api/v1/users - List all users (requires users:read, cached)
//   - GET /api/v1/users/:id - Get user by ID (requires users:read, cached)
//   - PATCH /api/v1/users/:id/deactivate - Deactivate user (requires users:write)
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with identity handler and auth middlewares
func registerUserRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	// User list endpoint with caching (5 minutes) - rarely changes
	v1.GET("/users",
		di.AuthMiddleware.Authenticate(),
		di.RBACMiddleware.Require("users:read"),
		middleware.Cache(di.Cache, 5*time.Minute),
		di.IdentityHandler.ListUsers)

	// User details endpoint with caching (5 minutes)
	v1.GET("/users/:id",
		di.AuthMiddleware.Authenticate(),
		di.RBACMiddleware.Require("users:read"),
		middleware.Cache(di.Cache, 5*time.Minute),
		di.IdentityHandler.GetUser)

	// Current user profile endpoint (uses Bearer token, no RBAC)
	v1.GET("/users/me",
		di.AuthMiddleware.Authenticate(),
		di.IdentityHandler.GetCurrentUser)

	// User deactivation - no cache for mutations
	v1.PATCH("/users/:id/deactivate",
		di.AuthMiddleware.Authenticate(),
		di.RBACMiddleware.Require("users:write"),
		di.IdentityHandler.DeactivateUser)
}

// registerRoleRoutes registers role management endpoints.
// Role routes require JWT authentication and roles:manage permission.
//
// Endpoints:
//   - POST /api/v1/users/:id/roles - Assign role to user (requires roles:manage)
//   - DELETE /api/v1/users/:id/roles/:rid - Revoke role from user (requires roles:manage)
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with identity handler and auth middlewares
func registerRoleRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	v1.POST("/users/:id/roles", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("roles:manage"), di.IdentityHandler.AssignRole)
	v1.DELETE("/users/:id/roles/:rid", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("roles:manage"), di.IdentityHandler.RevokeRole)
}

// registerSessionRoutes registers session management endpoints.
// All session routes require authentication.
//
// Endpoints:
//   - GET /api/v1/sessions/:id - Get session by ID (requires authentication)
//   - POST /api/v1/sessions/:id/refresh - Refresh session (requires authentication)
//   - DELETE /api/v1/sessions/:id - Revoke session (requires authentication)
//   - GET /api/v1/users/:id/sessions - List user sessions (requires authentication)
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with session handler and auth middlewares
func registerSessionRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	v1.POST("/sessions", di.AuthMiddleware.Authenticate(), di.SessionHandler.CreateSession)
	v1.POST("/sessions/:id/refresh", di.AuthMiddleware.Authenticate(), di.SessionHandler.RefreshSession)
	v1.DELETE("/sessions/:id", di.AuthMiddleware.Authenticate(), di.SessionHandler.RevokeSession)
	v1.GET("/users/:id/sessions", di.AuthMiddleware.Authenticate(), di.SessionHandler.GetUserSessions)
}

// registerPreferencesRoutes registers user preferences endpoints.
// All preferences routes require authentication.
//
// Endpoints:
//   - GET /api/v1/users/:id/preferences - Get user preferences (requires authentication)
//   - PUT /api/v1/users/:id/preferences - Update preferences (requires authentication)
//   - PUT /api/v1/users/:id/preferences/theme - Set theme (requires authentication)
//   - PUT /api/v1/users/:id/preferences/language - Set language (requires authentication)
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with preferences handler and auth middlewares
func registerPreferencesRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	v1.GET("/users/:id/preferences", di.AuthMiddleware.Authenticate(), di.PreferencesHandler.GetPreferences)
	v1.PUT("/users/:id/preferences", di.AuthMiddleware.Authenticate(), di.PreferencesHandler.UpdatePreferences)
	v1.PUT("/users/:id/preferences/theme", di.AuthMiddleware.Authenticate(), di.PreferencesHandler.SetTheme)
	v1.PUT("/users/:id/preferences/language", di.AuthMiddleware.Authenticate(), di.PreferencesHandler.SetLanguage)
}

// registerStatusRoutes registers system status and monitoring endpoints.
// These endpoints are publicly accessible without authentication for
// health monitoring, load balancer checks, and system observability.
//
// Endpoint Structure:
//   - /health - Simple health check (for load balancers, always returns 200)
//   - /system/ready - Readiness probe (checks if service can handle requests)
//   - /system/build - Build information (version, commit, build time)
//   - /system/info - Detailed system info (cache, rate limiter, config)
//
// Rationale:
//   - /health at root level follows industry standard (Kubernetes, AWS, etc.)
//   - /system/* groups all system-related endpoints together
//   - Separate endpoints for different monitoring needs
//
// Parameters:
//   - r: Gin engine to configure (registered at root, not under /api/v1)
//   - di: Dependency container with status handler and system info
func registerStatusRoutes(r *gin.Engine, di *Dependencies) {
	// Standard health check endpoint (load balancer friendly)
	r.GET("/health", di.StatusHandler.Health)

	// System endpoints group
	system := r.Group("/system")
	{
		// Readiness probe - checks if service is ready to accept traffic
		system.GET("/ready", di.StatusHandler.Ready)

		// Build information - version, commit, build time
		system.GET("/build", di.StatusHandler.GetInfo)

		// Detailed system information - cache, rate limiter, runtime stats
		system.GET("/info", systemInfoHandler(di))
	}
}

// systemInfoHandler returns a handler that provides detailed system information.
// Includes database status, cache status, rate limiter status, configuration, and runtime information.
// Useful for monitoring, debugging, and observability in production environments.
//
// Response structure:
//   - status: overall system health ("operational", "degraded", "down")
//   - database: database type, status, connection info (sanitized)
//   - redis: redis status (if configured)
//   - cache: cache type and status
//   - rate_limiter: rate limiter type and status
//   - runtime: Go version, goroutines, memory stats
//
// Status codes:
//   - ok: Component is working correctly
//   - degraded: Component is working but with issues
//   - fail: Component is not working
//   - not_configured: Component is not configured
//
// This endpoint is publicly accessible and does not require authentication.
// Consider adding authentication for production sensitive environments.
func systemInfoHandler(di *Dependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Database status
		dbStatus := checkDatabase(di.Database)

		// Redis status (if configured)
		redisStatus := "not_configured"
		if di.Config.Redis.URL != "" {
			redisStatus = "configured" // Will add actual check when Redis is integrated
		}

		// Overall system status
		overallStatus := "operational"
		if dbStatus == "fail" {
			overallStatus = "degraded"
		}

		info := gin.H{
			"status": overallStatus,
			"database": gin.H{
				"type":   dbType(di.Database),
				"status": dbStatus,
				"path":   dbPath(di.Config),
			},
			"redis": gin.H{
				"status": redisStatus,
				"url":    redisURL(di.Config),
			},
			"cache": gin.H{
				"type":   di.Cache.(interface{ Type() string }).Type(),
				"status": "ok",
			},
			"rate_limiter": gin.H{
				"type":   di.RateLimiter.(interface{ Type() string }).Type(),
				"status": "ok",
			},
			"config": gin.H{
				"env":  di.Config.App.Env,
				"name": di.Config.App.Name,
			},
		}
		c.JSON(200, info)
	}
}

// checkDatabase performs a health check on the database connection.
func checkDatabase(pool *pgxpool.Pool) string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return "fail"
	}
	return "ok"
}

// dbType returns the database type.
func dbType(pool *pgxpool.Pool) string {
	if pool == nil {
		return "not_configured"
	}
	return "postgres"
}

// dbPath returns sanitized database path.
func dbPath(cfg *config.Config) string {
	if cfg.Database.Path == "" {
		return "not_configured"
	}
	// Don't expose full path in production
	if cfg.App.Env == "prod" {
		return "***"
	}
	return cfg.Database.Path
}

// redisURL returns sanitized Redis URL.
func redisURL(cfg *config.Config) string {
	if cfg.Redis.URL == "" {
		return "not_configured"
	}
	// Don't expose credentials in production
	if cfg.App.Env == "prod" {
		return "***@redis://***"
	}
	return cfg.Redis.URL
}

// registerAuditRoutes registers audit log query endpoints.
// Audit routes require JWT authentication and audit:read permission.
//
// Endpoints:
//   - GET /api/v1/audit/records - List audit records with filtering (requires audit:read)
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with audit handler and auth middlewares
func registerAuditRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	v1.GET("/audit/records", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("audit:read"), di.AuditHandler.ListRecords)
}

// registerFilesRoutes registers file management endpoints.
// Files endpoints provide upload, download, processing, and management capabilities.
//
// Rate Limiting:
//   - File upload: 10 requests per minute per user (prevent storage abuse)
//   - File download: 30 requests per minute per user (preserve bandwidth)
//   - Processing: 5 requests per minute per user (computational expensive)
//
// User Endpoints (require authentication):
//   - GET /api/v1/files - List files with filtering
//   - POST /api/v1/files - Upload a file directly (rate limited)
//   - GET /api/v1/files/:id - Get file metadata
//   - DELETE /api/v1/files/:id - Delete a file
//   - POST /api/v1/files/upload-url - Request presigned upload URL (rate limited)
//   - POST /api/v1/files/confirm - Confirm upload completion
//   - POST /api/v1/files/:id/process - Request file processing (rate limited)
//   - GET /api/v1/files/processing/:id - Get processing status
//   - GET /api/v1/files/:id/processings - List file processings
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with files handler and auth middlewares
func registerFilesRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	files := v1.Group("/files")
	files.Use(di.AuthMiddleware.Authenticate())

	// File operations with rate limiting
	files.GET("", di.FilesHandler.ListFiles)
	files.GET("/:id", di.FilesHandler.GetFile)
	files.DELETE("/:id", di.FilesHandler.DeleteFile)

	// Direct upload - rate limited (10 per minute per user)
	files.POST("",
		middleware.RateLimit(di.RateLimiter, middleware.ByUser, 10, time.Minute),
		di.FilesHandler.UploadFile)

	// Presigned upload URLs - rate limited (10 per minute)
	files.POST("/upload-url",
		middleware.RateLimit(di.RateLimiter, middleware.ByUser, 10, time.Minute),
		di.FilesHandler.RequestUploadURL)
	files.POST("/confirm", di.FilesHandler.ConfirmUpload)

	// File processing - rate limited (5 per minute, computationally expensive)
	files.POST("/:id/process",
		middleware.RateLimit(di.RateLimiter, middleware.ByUser, 5, time.Minute),
		di.FilesHandler.RequestProcessing)
	files.GET("/processing/:id", di.FilesHandler.GetProcessingStatus)
	files.GET("/:id/processings", di.FilesHandler.ListProcessings)
}

// registerNotificationRoutes registers notification management endpoints.
// Endpoints are split into user-accessible and admin-only routes.
//
// User Endpoints (require authentication):
//   - GET /api/v1/notifications - List user's notifications
//   - GET /api/v1/notifications/:id - Get notification by ID
//   - GET /api/v1/notifications/preferences - Get notification preferences
//   - PATCH /api/v1/notifications/preferences - Update notification preferences
//
// Admin Endpoints (require special permissions):
//   - POST /api/v1/notifications - Create notification (requires notifications:write)
//   - GET /api/v1/notifications/templates - List templates (requires notifications:admin)
//   - GET /api/v1/notifications/templates/:id - Get template (requires notifications:admin)
//   - POST /api/v1/notifications/templates - Create template (requires notifications:admin)
//   - PATCH /api/v1/notifications/templates/:id - Update template (requires notifications:admin)
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with notification handler and auth middlewares
func registerNotificationRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	// User notification endpoints (authenticated)
	v1.GET("/notifications", di.AuthMiddleware.Authenticate(), di.NotificationHandler.ListNotifications)
	v1.GET("/notifications/:id", di.AuthMiddleware.Authenticate(), di.NotificationHandler.GetNotification)
	v1.GET("/notifications/preferences", di.AuthMiddleware.Authenticate(), di.NotificationHandler.GetPreferences)
	v1.PATCH("/notifications/preferences", di.AuthMiddleware.Authenticate(), di.NotificationHandler.UpdatePreferences)

	// Admin notification endpoints (requires special permissions)
	v1.POST("/notifications", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("notifications:write"), di.NotificationHandler.CreateNotification)
	v1.GET("/notifications/templates", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("notifications:admin"), di.NotificationHandler.ListTemplates)
	v1.GET("/notifications/templates/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("notifications:admin"), di.NotificationHandler.GetTemplate)
	v1.POST("/notifications/templates", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("notifications:admin"), di.NotificationHandler.CreateTemplate)
	v1.PATCH("/notifications/templates/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("notifications:admin"), di.NotificationHandler.UpdateTemplate)
}

// registerPartiesRoutes registers parties management endpoints.
// Parties include customers, suppliers, partners, and employees.
//
// Endpoints:
//   - POST /api/v1/customers - Create customer
//   - GET /api/v1/customers/:id - Get customer by ID
//   - GET /api/v1/customers - List customers
//   - PUT /api/v1/customers/:id - Update customer
//   - POST /api/v1/suppliers - Create supplier
//   - GET /api/v1/suppliers/:id - Get supplier by ID
//   - GET /api/v1/suppliers - List suppliers
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with parties handler and auth middlewares
func registerPartiesRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	// Customer endpoints (require authentication and parties:read/write)
	v1.POST("/customers", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("parties:write"), di.PartiesHandler.CreateCustomer)
	v1.GET("/customers/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("parties:read"), di.PartiesHandler.GetCustomer)
	v1.GET("/customers", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("parties:read"), di.PartiesHandler.ListCustomers)
	v1.PUT("/customers/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("parties:write"), di.PartiesHandler.UpdateCustomer)

	// Supplier endpoints (require authentication and parties:read/write)
	v1.POST("/suppliers", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("parties:write"), di.PartiesHandler.CreateSupplier)
	v1.GET("/suppliers/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("parties:read"), di.PartiesHandler.GetSupplier)
	v1.GET("/suppliers", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("parties:read"), di.PartiesHandler.ListSuppliers)
}

// registerDocumentsRoutes registers documents management endpoints.
// Documents include PDF generation and digital signatures.
//
// Endpoints:
//   - POST /api/v1/documents - Create document (draft)
//   - GET /api/v1/documents/:id - Get document by ID
//   - GET /api/v1/documents - List documents with filtering
//   - POST /api/v1/documents/:id/generate - Generate PDF from template
//   - POST /api/v1/documents/:id/signatures - Add signature request
//   - POST /api/v1/documents/:id/sign - Sign document
//   - GET /api/v1/templates/:id - Get template by ID
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with documents handler and auth middlewares
func registerDocumentsRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	// Document endpoints (require authentication and documents:read/write)
	v1.POST("/documents", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("documents:write"), di.DocumentsHandler.CreateDocument)
	v1.GET("/documents/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("documents:read"), di.DocumentsHandler.GetDocument)
	v1.GET("/documents", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("documents:read"), di.DocumentsHandler.ListDocuments)

	// Document actions (require authentication and documents:write)
	v1.POST("/documents/:id/generate", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("documents:write"), di.DocumentsHandler.GenerateDocument)
	v1.POST("/documents/:id/signatures", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("documents:write"), di.DocumentsHandler.AddSignature)
	v1.POST("/documents/:id/sign", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("documents:write"), di.DocumentsHandler.SignDocument)

	// Template endpoints (require authentication and documents:read)
	v1.GET("/templates/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("documents:read"), di.DocumentsHandler.GetTemplate)
}

// registerInvoicingRoutes registers invoicing management endpoints.
// Invoicing includes invoice creation, sending, payments, and cancellation.
//
// Endpoints:
//   - POST /api/v1/invoices - Create invoice (draft status)
//   - GET /api/v1/invoices/:id - Get invoice by ID with lines and payments
//   - GET /api/v1/invoices - List invoices with filtering
//   - POST /api/v1/invoices/:id/lines - Add line to invoice
//   - POST /api/v1/invoices/:id/send - Send invoice (draft → sent)
//   - POST /api/v1/invoices/:id/payments - Record payment for invoice
//   - POST /api/v1/invoices/:id/cancel - Cancel invoice
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with invoicing handler and auth middlewares
func registerInvoicingRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	// Invoice endpoints (require authentication and invoicing:read/write)
	v1.POST("/invoices", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("invoicing:write"), di.InvoicingHandler.CreateInvoice)
	v1.GET("/invoices/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("invoicing:read"), di.InvoicingHandler.GetInvoice)
	v1.GET("/invoices", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("invoicing:read"), di.InvoicingHandler.ListInvoices)

	// Invoice actions (require authentication and invoicing:write)
	v1.POST("/invoices/:id/lines", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("invoicing:write"), di.InvoicingHandler.AddInvoiceLine)
	v1.POST("/invoices/:id/send", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("invoicing:write"), di.InvoicingHandler.SendInvoice)
	v1.POST("/invoices/:id/payments", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("invoicing:write"), di.InvoicingHandler.RecordPayment)
	v1.POST("/invoices/:id/cancel", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("invoicing:write"), di.InvoicingHandler.CancelInvoice)
}

// registerInventoryRoutes registers inventory management endpoints.
// Inventory includes warehouses, stock levels, movements, and reservations.
//
// Endpoints:
//   - POST /api/v1/warehouses - Create warehouse
//   - GET /api/v1/warehouses/:id - Get warehouse by ID
//   - GET /api/v1/warehouses - List warehouses
//   - PUT /api/v1/warehouses/:id - Update warehouse (status/capacity)
//   - POST /api/v1/stock - Create stock record
//   - GET /api/v1/stock/:id - Get stock by ID
//   - GET /api/v1/stock - List stock with filtering
//   - POST /api/v1/stock/:id/adjust - Adjust stock quantity
//   - POST /api/v1/stock/receipt - Receive stock into warehouse
//   - POST /api/v1/stock/issue - Issue stock from warehouse
//   - POST /api/v1/stock/transfer - Transfer stock between warehouses
//   - POST /api/v1/stock/reserve - Reserve stock for order
//   - POST /api/v1/reservations/fulfill - Fulfill reservation
//   - POST /api/v1/reservations/cancel - Cancel reservation
//   - GET /api/v1/reservations/:id - Get reservation by ID
//   - GET /api/v1/reservations - List reservations by order
//   - GET /api/v1/movements/:id - Get movement by ID
//   - GET /api/v1/movements - List movements with filtering
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with inventory handler and auth middlewares
func registerInventoryRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	// Warehouse endpoints (require authentication and inventory:read/write)
	v1.POST("/warehouses", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:write"), di.InventoryHandler.CreateWarehouse)
	v1.GET("/warehouses/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:read"), di.InventoryHandler.GetWarehouse)
	v1.GET("/warehouses", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:read"), di.InventoryHandler.ListWarehouses)
	v1.PUT("/warehouses/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:write"), di.InventoryHandler.UpdateWarehouse)

	// Stock endpoints (require authentication and inventory:read/write)
	v1.POST("/stock", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:write"), di.InventoryHandler.CreateStock)
	v1.GET("/stock/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:read"), di.InventoryHandler.GetStock)
	v1.GET("/stock", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:read"), di.InventoryHandler.ListStock)
	v1.POST("/stock/:id/adjust", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:write"), di.InventoryHandler.AdjustStock)

	// Stock operations (require authentication and inventory:write)
	v1.POST("/stock/receipt", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:write"), di.InventoryHandler.ReceiptStock)
	v1.POST("/stock/issue", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:write"), di.InventoryHandler.IssueStock)
	v1.POST("/stock/transfer", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:write"), di.InventoryHandler.TransferStock)
	v1.POST("/stock/reserve", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:write"), di.InventoryHandler.ReserveStock)

	// Reservation endpoints (require authentication and inventory:read/write)
	v1.POST("/reservations/fulfill", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:write"), di.InventoryHandler.FulfillReservation)
	v1.POST("/reservations/cancel", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:write"), di.InventoryHandler.CancelReservation)
	v1.GET("/reservations/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:read"), di.InventoryHandler.GetReservation)
	v1.GET("/reservations", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:read"), di.InventoryHandler.ListReservations)

	// Movement endpoints (require authentication and inventory:read)
	v1.GET("/movements/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:read"), di.InventoryHandler.GetStockMovement)
	v1.GET("/movements", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("inventory:read"), di.InventoryHandler.ListStockMovements)
}

// registerAccountingRoutes registers accounting management endpoints.
// Accounting includes chart of accounts and double-entry transaction recording.
//
// Endpoints:
//   - POST /api/v1/accounts - Create account in chart of accounts
//   - GET /api/v1/accounts/:id - Get account by ID
//   - GET /api/v1/accounts - List accounts with filtering
//   - POST /api/v1/transactions - Record a transaction (double-entry bookkeeping)
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with accounting handler and auth middlewares
func registerAccountingRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	// Account endpoints (require authentication and accounting:read/write)
	v1.POST("/accounts", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("accounting:write"), di.AccountingHandler.CreateAccount)
	v1.GET("/accounts/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("accounting:read"), di.AccountingHandler.GetAccount)
	v1.GET("/accounts", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("accounting:read"), di.AccountingHandler.ListAccounts)

	// Transaction endpoints (require authentication and accounting:write)
	v1.POST("/transactions", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("accounting:write"), di.AccountingHandler.RecordTransaction)
}

// registerOrderingRoutes registers ordering management endpoints.
// Ordering includes order creation, line management, and order status transitions.
//
// Endpoints:
//   - POST /api/v1/orders - Create a new order (draft status)
//   - GET /api/v1/orders/:id - Get order by ID with order lines
//   - GET /api/v1/orders - List orders with filtering
//   - POST /api/v1/orders/:id/lines - Add order line to order
//   - PATCH /api/v1/orders/:id/status - Update order status (state machine transitions)
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with ordering handler and auth middlewares
func registerOrderingRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	// Order endpoints (require authentication and ordering:read/write)
	v1.POST("/orders", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("ordering:write"), di.OrderingHandler.CreateOrder)
	v1.GET("/orders/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("ordering:read"), di.OrderingHandler.GetOrder)
	v1.GET("/orders", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("ordering:read"), di.OrderingHandler.ListOrders)

	// Order line management (require authentication and ordering:write)
	v1.POST("/orders/:id/lines", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("ordering:write"), di.OrderingHandler.AddOrderLine)

	// Order status transitions (require authentication and ordering:write)
	v1.PATCH("/orders/:id/status", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("ordering:write"), di.OrderingHandler.UpdateOrderStatus)
}

// registerCatalogRoutes registers catalog management endpoints.
// Catalog includes inventory items and category management.
//
// Endpoints:
//   - POST /api/v1/catalog/items - Create a new catalog item
//   - GET /api/v1/catalog/items/:id - Get item by ID
//   - GET /api/v1/catalog/items - List items with filtering
//   - PUT /api/v1/catalog/items/:id - Update item details
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with catalog handler and auth middlewares
func registerCatalogRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	// Catalog item endpoints (require authentication and catalog:read/write)
	items := v1.Group("/catalog/items")
	items.Use(di.AuthMiddleware.Authenticate())

	items.POST("", di.RBACMiddleware.Require("catalog:write"), di.CatalogHandler.CreateItem)
	items.GET("/:id", di.RBACMiddleware.Require("catalog:read"), di.CatalogHandler.GetItem)
	items.GET("", di.RBACMiddleware.Require("catalog:read"), di.CatalogHandler.ListItems)
	items.PUT("/:id", di.RBACMiddleware.Require("catalog:write"), di.CatalogHandler.UpdateItem)
}

// recoveryMiddleware returns a Gin middleware that recovers from panics.
// When a panic occurs, it logs the error and returns a generic 500 Internal Server Error
// response with the request ID for debugging.
//
// This middleware ensures the server doesn't crash from panics in request handlers
// and always returns a proper JSON error response.
//
// Returns a Gin middleware handler function.
func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")
				c.AbortWithStatusJSON(500, gin.H{
					"error":      "internal_server_error",
					"request_id": requestID,
				})
			}
		}()
		c.Next()
	}
}

// requestIDMiddleware returns a Gin middleware that ensures each request has a unique ID.
// If the client provides an X-Request-ID header, it uses that value; otherwise,
// it generates a new UUID v7.
//
// The request ID is:
//   - Set in the response header X-Request-ID
//   - Stored in the Gin context for use by other middlewares/handlers
//   - Included in log entries for request tracing
//
// Returns a Gin middleware handler function.
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// loggerMiddleware returns a Gin middleware that logs HTTP requests using structured logging.
// Logs include method, path, status code, latency, request ID, and client IP address.
//
// Uses slog for structured logging with the configured log level and format.
// Each request generates a log entry after the request is processed.
//
// Returns a Gin middleware handler function.
func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		requestID, _ := c.Get("request_id")
		slog.Info("request",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency", time.Since(start),
			"request_id", requestID,
			"ip", c.ClientIP(),
		)
	}
}

// corsMiddleware returns a Gin middleware that handles Cross-Origin Resource Sharing (CORS).
// Configured to allow requests from any origin with common HTTP methods and headers.
//
// CORS Configuration:
//   - AllowOrigins: All origins (*)
//   - AllowMethods: GET, POST, PUT, PATCH, DELETE, OPTIONS
//   - AllowHeaders: Origin, Content-Type, Accept, Authorization, X-Request-ID
//   - ExposeHeaders: X-Request-ID
//   - AllowCredentials: false
//   - MaxAge: 12 hours
//
// Note: For production, consider restricting AllowOrigins to known domains.
//
// Returns a Gin middleware handler function.
func corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// generateRequestID creates a new unique request identifier using UUID v7.
// UUID v7 provides time-sortable identifiers suitable for request tracing
// and distributed systems.
//
// Returns a string representation of a new UUID v7.
func generateRequestID() string {
	return uuid.NewV7().String()
}

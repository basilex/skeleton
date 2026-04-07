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
//   - Status routes (/health, /build): Health checks and build information
//
// Middleware chain (applied in order):
//  1. Recovery middleware - Panic recovery and error response
//  2. Request ID middleware - Unique request identification
//  3. Logger middleware - Structured request logging
//  4. CORS middleware - Cross-origin resource sharing
//
// Protected routes use authentication and RBAC middlewares as needed.
package main

import (
	"log/slog"
	"time"

	"github.com/basilex/skeleton/pkg/config"
	"github.com/basilex/skeleton/pkg/uuid"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
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
// Routes are grouped under /api/v1 for versioned endpoints, with status
// endpoints at the root level for health checks.
//
// Route Structure:
//
//	/api/v1/auth -> Authentication endpoints
//	/api/v1/users -> User management endpoints
//	/api/v1/audit -> Audit log endpoints
//	/api/v1/notifications -> Notification endpoints
//	/health -> Health check endpoint
//	/build -> Build information endpoint
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
		registerAuditRoutes(v1, di)
		registerNotificationRoutes(v1, di)
	}

	registerStatusRoutes(r, di)
}

// registerAuthRoutes registers authentication-related endpoints.
// All auth routes are publicly accessible except /auth/logout and /auth/me
// which require session authentication.
//
// Endpoints:
//   - POST /api/v1/auth/register - Register a new user
//   - POST /api/v1/auth/login - Authenticate user and create session
//   - POST /api/v1/auth/refresh - Refresh access token
//   - POST /api/v1/auth/logout - End user session (requires authentication)
//   - GET /api/v1/auth/me - Get current user profile (requires authentication)
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with identity handler and session middleware
func registerAuthRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	v1.POST("/auth/register", di.IdentityHandler.Register)
	v1.POST("/auth/login", di.IdentityHandler.Login)
	v1.POST("/auth/refresh", di.IdentityHandler.Refresh)
	v1.POST("/auth/logout", di.SessionMiddleware.Authenticate(), di.IdentityHandler.Logout)
	v1.GET("/auth/me", di.SessionMiddleware.Authenticate(), di.IdentityHandler.GetMyProfile)
}

// registerUserRoutes registers user management endpoints.
// All user routes require JWT authentication and RBAC permissions.
//
// Endpoints:
//   - GET /api/v1/users - List all users (requires users:read)
//   - GET /api/v1/users/:id - Get user by ID (requires users:read)
//   - PATCH /api/v1/users/:id/deactivate - Deactivate user (requires users:write)
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with identity handler and auth middlewares
func registerUserRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	v1.GET("/users", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("users:read"), di.IdentityHandler.ListUsers)
	v1.GET("/users/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("users:read"), di.IdentityHandler.GetUser)
	v1.PATCH("/users/:id/deactivate", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("users:write"), di.IdentityHandler.DeactivateUser)
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

// registerStatusRoutes registers system status endpoints.
// These endpoints are publicly accessible without authentication for
// health monitoring and build information display.
//
// Endpoints:
//   - GET /health - Health check endpoint (public)
//   - GET /build - Build information endpoint (public)
//
// Parameters:
//   - r: Gin engine to configure (registered at root, not under /api/v1)
//   - di: Dependency container with status handler
func registerStatusRoutes(r *gin.Engine, di *Dependencies) {
	r.GET("/health", di.StatusHandler.Health)
	r.GET("/build", di.StatusHandler.GetInfo)
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
// User Endpoints (require authentication):
//   - GET /api/v1/files - List files with filtering
//   - POST /api/v1/files - Upload a file directly (small files <5MB)
//   - GET /api/v1/files/:id - Get file metadata
//   - DELETE /api/v1/files/:id - Delete a file
//   - POST /api/v1/files/upload-url - Request presigned upload URL (large files)
//   - POST /api/v1/files/confirm - Confirm upload completion
//   - POST /api/v1/files/:id/process - Request file processing
//   - GET /api/v1/files/processing/:id - Get processing status
//   - GET /api/v1/files/:id/processings - List file processings
//
// Parameters:
//   - v1: Router group for /api/v1 endpoints
//   - di: Dependency container with files handler and auth middlewares
func registerFilesRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	files := v1.Group("/files")
	files.Use(di.AuthMiddleware.Authenticate())

	// File CRUD operations
	files.GET("", di.FilesHandler.ListFiles)
	files.GET("/:id", di.FilesHandler.GetFile)
	files.DELETE("/:id", di.FilesHandler.DeleteFile)

	// Direct upload (small files)
	files.POST("", di.FilesHandler.UploadFile)

	// Presigned upload URLs (large files)
	files.POST("/upload-url", di.FilesHandler.RequestUploadURL)
	files.POST("/confirm", di.FilesHandler.ConfirmUpload)

	// File processing
	files.POST("/:id/process", di.FilesHandler.RequestProcessing)
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
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
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

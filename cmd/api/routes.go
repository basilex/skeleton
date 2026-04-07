package main

import (
	"log/slog"
	"time"

	"github.com/basilex/skeleton/pkg/config"
	"github.com/basilex/skeleton/pkg/uuid"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func setupRouter(cfg *config.Config, di *Dependencies) *gin.Engine {
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	registerGlobalMiddleware(r)
	registerRoutes(r, di)

	return r
}

func registerGlobalMiddleware(r *gin.Engine) {
	r.Use(recoveryMiddleware())
	r.Use(requestIDMiddleware())
	r.Use(loggerMiddleware())
	r.Use(corsMiddleware())
}

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

func registerAuthRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	v1.POST("/auth/register", di.IdentityHandler.Register)
	v1.POST("/auth/login", di.IdentityHandler.Login)
	v1.POST("/auth/refresh", di.IdentityHandler.Refresh)
	v1.POST("/auth/logout", di.SessionMiddleware.Authenticate(), di.IdentityHandler.Logout)
	v1.GET("/auth/me", di.SessionMiddleware.Authenticate(), di.IdentityHandler.GetMyProfile)
}

func registerUserRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	v1.GET("/users", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("users:read"), di.IdentityHandler.ListUsers)
	v1.GET("/users/:id", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("users:read"), di.IdentityHandler.GetUser)
	v1.PATCH("/users/:id/deactivate", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("users:write"), di.IdentityHandler.DeactivateUser)
}

func registerRoleRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	v1.POST("/users/:id/roles", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("roles:manage"), di.IdentityHandler.AssignRole)
	v1.DELETE("/users/:id/roles/:rid", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("roles:manage"), di.IdentityHandler.RevokeRole)
}

func registerStatusRoutes(r *gin.Engine, di *Dependencies) {
	r.GET("/health", di.StatusHandler.Health)
	r.GET("/build", di.StatusHandler.GetInfo)
}

func registerAuditRoutes(v1 *gin.RouterGroup, di *Dependencies) {
	v1.GET("/audit/records", di.AuthMiddleware.Authenticate(), di.RBACMiddleware.Require("audit:read"), di.AuditHandler.ListRecords)
}

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

func generateRequestID() string {
	return uuid.NewV7().String()
}

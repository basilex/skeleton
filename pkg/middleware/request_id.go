// Package middleware provides HTTP middleware components for the Gin framework.
// It includes request ID generation, logging, and panic recovery middleware.
package middleware

import (
	"github.com/basilex/skeleton/pkg/uuid"
	"github.com/gin-gonic/gin"
)

// RequestID returns a Gin middleware that ensures each request has a unique X-Request-ID header.
// If the header is already present, it uses the existing value. Otherwise, it generates a new UUID v7.
// The request ID is also stored in the Gin context for use by other middleware and handlers.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewV7().String()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

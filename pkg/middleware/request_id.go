package middleware

import (
	"github.com/basilex/skeleton/pkg/uuid"
	"github.com/gin-gonic/gin"
)

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

package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")
				slog.Error("panic recovered",
					"error", err,
					"request_id", requestID,
				)
				c.AbortWithStatusJSON(500, gin.H{
					"error":      "internal_server_error",
					"request_id": requestID,
				})
			}
		}()
		c.Next()
	}
}

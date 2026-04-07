package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

// Recovery returns a Gin middleware that recovers from panics in request handlers.
// It logs the panic details with the request ID and returns a 500 Internal Server Error response.
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

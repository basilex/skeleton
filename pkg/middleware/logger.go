package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
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

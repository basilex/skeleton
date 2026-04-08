package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/basilex/skeleton/pkg/ratelimit"
	"github.com/gin-gonic/gin"
)

// RateLimit creates a rate limiting middleware.
func RateLimit(limiter ratelimit.Limiter, keyFunc KeyFunc, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate rate limit key
		key := keyFunc(c)

		// Check rate limit
		allowed, err := limiter.Allow(context.Background(), key)
		if err != nil {
			// On error, allow request (fail-open)
			c.Next()
			return
		}

		// Get remaining requests
		remaining, _ := limiter.Remaining(context.Background(), key)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(window).Unix(), 10))

		if !allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(window.Seconds()), 10))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// KeyFunc generates a rate limit key from the request.
type KeyFunc func(c *gin.Context) string

// ByIP generates a rate limit key based on client IP.
func ByIP(c *gin.Context) string {
	return fmt.Sprintf("ratelimit:ip:%s", c.ClientIP())
}

// ByUser generates a rate limit key based on authenticated user ID.
func ByUser(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ByIP(c)
	}
	return fmt.Sprintf("ratelimit:user:%s", userID)
}

// ByEndpoint generates a rate limit key based on endpoint and method.
func ByEndpoint(c *gin.Context) string {
	return fmt.Sprintf("ratelimit:endpoint:%s:%s", c.Request.Method, c.FullPath())
}

// ByUserAndEndpoint generates a rate limit key based on user and endpoint.
func ByUserAndEndpoint(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ByEndpoint(c)
	}
	return fmt.Sprintf("ratelimit:user:%s:endpoint:%s:%s", userID, c.Request.Method, c.FullPath())
}

// ByAPIKey generates a rate limit key based on API key.
func ByAPIKey(c *gin.Context) string {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		return ByIP(c)
	}
	return fmt.Sprintf("ratelimit:apikey:%s", apiKey)
}

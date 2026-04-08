package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/basilex/skeleton/pkg/cache"
	"github.com/gin-gonic/gin"
)

// CacheResponseWriter captures response for caching.
type CacheResponseWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *CacheResponseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

// Cache middleware caches HTTP responses.
func Cache(cacheStore cache.Cache, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Generate cache key
		key := cacheKey(c)

		// Try to get from cache
		var resp cachedResponse
		if err := cacheStore.Get(context.Background(), key, &resp); err == nil {
			// Cache hit
			for k, v := range resp.Headers {
				c.Writer.Header()[k] = v
			}
			c.Writer.Header().Set("X-Cache", "HIT")
			c.Data(resp.StatusCode, resp.ContentType, resp.Body)
			c.Abort()
			return
		}

		// Cache miss - wrap writer to capture response
		writer := &CacheResponseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0, 1024),
		}
		c.Writer = writer

		c.Next()

		// Only cache successful responses
		if c.Writer.Status() >= 200 && c.Writer.Status() < 400 {
			resp := cachedResponse{
				StatusCode:  c.Writer.Status(),
				ContentType: c.Writer.Header().Get("Content-Type"),
				Body:        writer.body,
				Headers:     make(map[string][]string),
			}

			// Copy relevant headers
			for k, v := range c.Writer.Header() {
				if shouldCacheHeader(k) {
					resp.Headers[k] = v
				}
			}

			// Store in cache
			_ = cacheStore.Set(context.Background(), key, resp, ttl)
		}
	}
}

type cachedResponse struct {
	StatusCode  int
	ContentType string
	Body        []byte
	Headers     map[string][]string
}

func cacheKey(c *gin.Context) string {
	h := sha256.New()
	h.Write([]byte(c.Request.Method))
	h.Write([]byte(c.Request.URL.Path))
	h.Write([]byte(c.Request.URL.RawQuery))

	// Include user context if authenticated
	if userID, exists := c.Get("user_id"); exists {
		h.Write([]byte(userID.(string)))
	}

	return "api:" + hex.EncodeToString(h.Sum(nil))
}

func shouldCacheHeader(header string) bool {
	// Don't cache these headers
	skipHeaders := map[string]bool{
		"Set-Cookie":    true,
		"X-Request-Id":  true,
		"X-Cache":       true,
		"Authorization": true,
	}

	return !skipHeaders[header]
}

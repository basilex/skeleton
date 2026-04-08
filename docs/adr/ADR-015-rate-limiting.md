# ADR-015: Rate Limiting Strategy

## Status

Accepted

## Context

APIs need rate limiting to:
- Prevent abuse and DDoS attacks
- Protect against brute force attacks
- Ensure fair resource distribution
- Control infrastructure costs
- Comply with API usage policies

Current state:
- No rate limiting exists
- APIs vulnerable to abuse
- No protection against brute force login attempts
- File uploads can overwhelm storage
- Expensive operations can be repeatedly called

## Decision

Implement a **multi-level rate limiting architecture** with pluggable algorithms:

### 1. Rate Limiter Interface

```go
type Limiter interface {
    // Allow checks if a request is allowed for the given key
    Allow(ctx context.Context, key string) (bool, error)
    
    // AllowN checks if N requests are allowed for the given key
    AllowN(ctx context.Context, key string, n int) (bool, error)
    
    // Remaining returns the number of requests remaining in the current window
    Remaining(ctx context.Context, key string) (int, error)
    
    // Reset resets the rate limit for the given key
    Reset(ctx context.Context, key string) error
    
    // Configure updates the rate limit configuration
    Configure(config RateLimitConfig) error
}

type RateLimitConfig struct {
    Limit      int           // Maximum requests allowed
    Window     time.Duration // Time window
    KeyPrefix  string        // Key prefix for namespacing
}
```

### 2. Implementations

#### Token Bucket (In-Memory)

```go
type TokenBucketLimiter struct {
    buckets sync.Map
    config  RateLimitConfig
}
```

- Fast local rate limiting
- No network overhead
- Suitable for single-instance deployments
- Good for development and testing
- Memory-efficient

**Algorithm:**
```
1. Each bucket has capacity C tokens
2. Tokens refill at rate R per second
3. Each request consumes 1 token
4. If bucket empty, request denied
```

#### Sliding Window (Redis-based)

```go
type SlidingWindowLimiter struct {
    client *redis.Client
    config RateLimitConfig
}
```

- Distributed rate limiting
- Consistent across all instances
- Precise counting
- Suitable for production
- Requires Redis

**Algorithm:**
```
1. Use Redis ZSET to store timestamps
2. Count requests in sliding window [now-window, now]
3. If count < limit, allow and add timestamp
4. Remove old timestamps outside window
```

### 3. Rate Limit Middleware

```go
func RateLimitMiddleware(limiter Limiter, keyFunc KeyFunc, limit int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := keyFunc(c)
        
        allowed, err := limiter.Allow(c.Request.Context(), key)
        if err != nil {
            // On error, allow request (fail-open)
            c.Next()
            return
        }
        
        if !allowed {
            c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
            c.Header("X-RateLimit-Remaining", "0")
            c.Header("X-RateLimit-Reset", resetTime)
            
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "rate_limit_exceeded",
                "message": "Too many requests. Please try again later.",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 4. Key Functions

Different strategies for generating rate limit keys:

```go
// Per-IP rate limiting
func ByIP(c *gin.Context) string {
    return fmt.Sprintf("ratelimit:ip:%s", c.ClientIP())
}

// Per-user rate limiting (requires authentication)
func ByUser(c *gin.Context) string {
    userID := getUserID(c)
    return fmt.Sprintf("ratelimit:user:%s", userID)
}

// Per-endpoint rate limiting
func ByEndpoint(c *gin.Context) string {
    return fmt.Sprintf("ratelimit:endpoint:%s:%s", c.Request.Method, c.FullPath())
}

// Combined: Per-user per-endpoint
func ByUserAndEndpoint(c *gin.Context) string {
    userID := getUserID(c)
    return fmt.Sprintf("ratelimit:user:%s:endpoint:%s:%s", 
        userID, c.Request.Method, c.FullPath())
}
```

### 5. Rate Limit Tiers

Recommended rate limits for different operations:

```go
// Authentication endpoints (protect against brute force)
authLimits := RateLimitConfig{
    Login:          {Limit: 5, Window: 1 * time.Minute},     // 5 per minute
    Register:       {Limit: 3, Window: 1 * time.Hour},       // 3 per hour
    Refresh:        {Limit: 30, Window: 1 * time.Minute},    // 30 per minute
    PasswordReset:  {Limit: 3, Window: 1 * time.Hour},       // 3 per hour
}

// File operations (protect storage)
fileLimits := RateLimitConfig{
    Upload:         {Limit: 10, Window: 1 * time.Minute},    // 10 per minute
    Download:       {Limit: 30, Window: 1 * time.Minute},   // 30 per minute
    Delete:         {Limit: 20, Window: 1 * time.Minute},   // 20 per minute
}

// API endpoints (fair usage)
apiLimits := RateLimitConfig{
    Global:         {Limit: 1000, Window: 1 * time.Minute},  // 1000 per minute
    PerUser:        {Limit: 100, Window: 1 * time.Minute},   // 100 per minute
    ReadOperations: {Limit: 200, Window: 1 * time.Minute},   // 200 per minute
    WriteOperations: {Limit: 50, Window: 1 * time.Minute},  // 50 per minute
}
```

### 6. Implementation Strategy

```go
// Factory pattern for creating limiters
func NewLimiter(config Config) Limiter {
    switch config.Type {
    case "memory":
        return NewTokenBucketLimiter(config.Memory)
    case "redis":
        return NewSlidingWindowLimiter(config.Redis)
    default:
        // Fallback chain: Redis -> Memory
        if redisClient != nil {
            return NewSlidingWindowLimiter(redisClient, config)
        }
        return NewTokenBucketLimiter(config)
    }
}
```

### 7. Fallback Strategy

```
Primary: Sliding Window (Redis)
    ↓ (if Redis unavailable)
Fallback: Token Bucket (Memory)
    ↓ (if rate limiter fails)
Fail-open: Allow request
```

## Use Cases

### 1. Authentication Endpoints

```go
// Protect login from brute force
auth.POST("/login",
    middleware.RateLimit(limiter, middleware.ByIP, 5, 1*time.Minute),
    handler.Login)

// Prevent mass registration
auth.POST("/register",
    middleware.RateLimit(limiter, middleware.ByIP, 3, 1*time.Hour),
    handler.Register)
```

### 2. File Uploads

```go
// Limit file uploads
files.POST("/upload",
    middleware.RateLimit(limiter, middleware.ByUser, 10, 1*time.Minute),
    handler.UploadFile)

// Limit expensive operations
files.POST("/:id/process",
    middleware.RateLimit(limiter, middleware.ByUser, 5, 1*time.Minute),
    handler.RequestProcessing)
```

### 3. API Endpoints

```go
// Global API rate limit
api.Use(middleware.RateLimit(limiter, middleware.ByIP, 1000, 1*time.Minute))

// Per-user rate limit
api.Use(middleware.RateLimit(limiter, middleware.ByUser, 100, 1*time.Minute))
```

### 4. Expensive Queries

```go
// Paginated list endpoints
api.GET("/users",
    middleware.RateLimit(limiter, middleware.ByUserAndEndpoint, 30, 1*time.Minute),
    handler.ListUsers)

// Search endpoints
api.GET("/search",
    middleware.RateLimit(limiter, middleware.ByUser, 20, 1*time.Minute),
    handler.Search)
```

## Configuration

```yaml
rate_limit:
  enabled: true
  type: sliding_window  # sliding_window | token_bucket
  fallback: token_bucket  # fallback if primary fails
  
  redis:
    key_prefix: "ratelimit"
    pool_size: 10
    
  defaults:
    global:
      limit: 1000
      window: 1m
    
  endpoints:
    auth/login:
      limit: 5
      window: 1m
      key: ip
    
    auth/register:
      limit: 3
      window: 1h
      key: ip
    
    files/upload:
      limit: 10
      window: 1m
      key: user
    
    users/list:
      limit: 30
      window: 1m
      key: user_and_endpoint
```

## Response Headers

Standard rate limit headers:

```
X-RateLimit-Limit: 100        # Maximum requests allowed
X-RateLimit-Remaining: 95     # Requests remaining in current window
X-RateLimit-Reset: 1640995200 # Unix timestamp when window resets
Retry-After: 60               # Seconds to wait before retry (on 429)
```

## Implementation Layers

### Layer 1: Core Interface (`pkg/ratelimit/limiter.go`)
- Limiter interface definition
- Rate limit errors
- Helper functions

### Layer 2: Implementations (`pkg/ratelimit/`)
- `token_bucket.go` - In-memory rate limiter
- `sliding_window.go` - Redis-based rate limiter

### Layer 3: Middleware (`pkg/middleware/ratelimit.go`)
- HTTP middleware
- Key functions (IP, user, endpoint)
- Header handling

### Layer 4: Integration (`cmd/api/`)
- Configure rate limiters
- Apply to routes
- Fallback setup

## Trade-offs

### Positive
- ✅ Protects against DDoS and brute force attacks
- ✅ Ensures fair resource distribution
- ✅ Prevents infrastructure overload
- ✅ Distributed rate limiting with Redis
- ✅ Graceful fallback to in-memory
- ✅ Per-user, per-IP, per-endpoint flexibility
- ✅ Fail-open strategy (availability over strict limiting)

### Negative
- ❌ Adds latency (Redis round-trip)
- ❌ Redis dependency for distributed limiting
- ❌ Potential false positives (shared IPs, NAT)
- ❌ Complexity in configuration
- ❌ Memory usage for in-memory fallback

### Neutral
- Token bucket is simpler but less accurate than sliding window
- Sliding window is more precise but requires Redis
- Rate limits are configurable and context-dependent

## Testing Strategy

### Unit Tests
- Mock Limiter interface
- Test token bucket algorithm
- Test sliding window algorithm
- Test key generation functions

### Integration Tests
- Redis-based rate limiting
- In-memory fallback
- Concurrent requests
- Window expiration

### Load Tests
- Rate limit under high load
- Redis connection failures
- Latency measurements
- Memory usage

## Security Considerations

1. **IP Spoofing**: Use `X-Forwarded-For` carefully, validate proxies
2. **Bot Detection**: Combine rate limiting with CAPTCHA for suspicious behavior
3. **Distributed Attacks**: Rate limit by user ID when authenticated
4. **DoS via Rate Limiting**: Implement circuit breakers
5. **Key Enumeration**: Rate limit by API key, not just IP

## Monitoring & Alerting

```go
// Metrics to track
- ratelimit_requests_total{limiter, endpoint, result}  // allowed|denied
- ratelimit_latency_seconds{limiter, operation}
- ratelimit_active_keys{limiter}
- ratelimit_redis_errors_total{error_type}
- ratelimit_fallback_activations_total{from, to}
```

## Alerting Rules

```yaml
# Alert when rate limiting kicks infrequently
- alert: RateLimitTriggered
  expr: rate(ratelimit_requests_total{result="denied"}[5m]) > 10
  severity: warning
  
# Alert when Redis failures spike
- alert: RateLimitRedisErrors
  expr: rate(ratelimit_redis_errors_total[5m]) > 0.1
  severity: critical
  
# Alert when fallback is frequently used
- alert: RateLimitFallbackActive
  expr: rate(ratelimit_fallback_activations_total[10m]) > 1
  severity: warning
```

## Migration Plan

### Phase 1: Infrastructure
1. Create `pkg/ratelimit` package
2. Implement Limiter interface
3. Implement Token Bucket limiter
4. Implement Sliding Window limiter
5. Add configuration support

### Phase 2: Integration
1. Add rate limit middleware
2. Apply to authentication endpoints
3. Apply to file operations
4. Apply to expensive queries

### Phase 3: Monitoring
1. Add metrics collection
2. Set up dashboards
3. Configure alerting
4. Monitor false positive rate

## Examples

### Example 1: Login Rate Limiting

```go
// Token bucket for development
limiter := ratelimit.NewTokenBucketLimiter(ratelimit.RateLimitConfig{
    Limit:     5,
    Window:    1 * time.Minute,
    KeyPrefix: "auth:login",
})

// Sliding window for production
limiter := ratelimit.NewSlidingWindowLimiter(redisClient, ratelimit.RateLimitConfig{
    Limit:     5,
    Window:    1 * time.Minute,
    KeyPrefix: "auth:login",
})

// Use in route
auth.POST("/login",
    middleware.RateLimit(limiter, middleware.ByIP, 5, 1*time.Minute),
    handler.Login)
```

### Example 2: User API Rate Limiting

```go
// Combined rate limits
api.Use(middleware.RateLimit(globalLimiter, middleware.ByIP, 1000, 1*time.Minute))
api.Use(middleware.RateLimit(userLimiter, middleware.ByUser, 100, 1*time.Minute))

// Specific endpoint limit
api.GET("/search",
    middleware.RateLimit(searchLimiter, middleware.ByUserAndEndpoint, 20, 1*time.Minute),
    handler.Search)
```

## References

- [Redis Rate Limiting](https://redis.io/glossary/rate-limiting/)
- [Token Bucket Algorithm](https://en.wikipedia.org/wiki/Token_bucket)
- [Sliding Window Algorithm](https://blog.cloudflare.com/counting-things-a-lot-of-different-things/)
- [Rate Limiting Best Practices](https://cloud.google.com/architecture/rate-limiting-strategies-techniques)
- [ADR-014: Caching Strategy](ADR-014-caching.md) - Uses same Redis infrastructure
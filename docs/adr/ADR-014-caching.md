# ADR-014: Caching Strategy

## Status

Accepted

## Context

Applications need caching to:
- Reduce database load
- Improve response times
- Handle high traffic efficiently
- Scale horizontally with multiple instances

Current state:
- No caching layer exists
- Database queries are repeated frequently
- User sessions, permissions, and file metadata are queried repeatedly
- API responses are regenerated for identical requests

## Decision

Implement a **multi-tier caching architecture** with pluggable backends:

### 1. Cache Interface

```go
type Cache interface {
    // Basic operations
    Get(ctx context.Context, key string, dest interface{}) error
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    
    // Bulk operations
    GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)
    SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
    DeleteMulti(ctx context.Context, keys []string) error
    
    // Pattern operations
    DeleteByPattern(ctx context.Context, pattern string) error
    
    // Utility
    Exists(ctx context.Context, key string) (bool, error)
    TTL(ctx context.Context, key string) (time.Duration, error)
    Clear(ctx context.Context) error
}
```

### 2. Implementations

#### Memory Cache(for Development)
```go
type MemoryCache struct {
    store sync.Map
    // TTL tracking with cleanup goroutine
}
```

- Fast local caching
- No external dependencies
- Good for development and testing
- Not suitable for multi-instance deployments

#### Redis Cache (for Production)
```go
type RedisCache struct {
    client *redis.Client
    prefix string
}
```

- Distributed caching
- Shared across all instances
- Persistence options
- Pattern-based invalidation

### 3. Cache Decorators

Repository decorators implement the cache-aside pattern:

```go
// Cache-aside: Check cache first, then database
func (d *CachedUserRepo) GetByID(ctx context.Context, id UserID) (*User, error) {
    key := fmt.Sprintf("user:%s", id)
    
    // Try cache first
    var user User
    if err := d.cache.Get(ctx, key, &user); err == nil {
        return &user, nil
    }
    
    // Cache miss - get from database
    user, err := d.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    _ = d.cache.Set(ctx, key, user, d.ttl)
    
    return user, nil
}
```

### 4. HTTP Cache Middleware

```go
// Cache HTTP responses by path/query
func CacheMiddleware(cache Cache, ttl time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Generate cache key from request
        key := cacheKey(c.Request)
        
        // Try cache
        var response CachedResponse
        if err := cache.Get(c, key, &response); err == nil {
            c.JSON(response.Status, response.Body)
            return
        }
        
        // Cache miss - continue to handler
        c.Next()
        
        // Store successful responses
        if c.Writer.Status() < 400 {
            _ = cache.Set(c, key, response, ttl)
        }
    }
}
```

### 5. Cache Keys Structure

```
{context}:{entity}:{id}:{field}

Examples:
- identity:user:0192e5c8-7f0b-7d2e-8b1a-5c3e2d1f0a9b:profile
- identity:session:abc123:data
- identity:role:admin:permissions
- files:file:0192e5c9-...:metadata
- api:users:list:page:1:limit:20
```

### 6. TTL Strategy

Different TTLs for different data types:

```go
const (
    UserSessionTTL      = 15 * time.Minute  // Short-lived, frequently accessed
    UserProfileTTL      = 1 * time.Hour     // Medium-lived, less frequent changes
    RolePermissionsTTL  = 2 * time.Hour     // Long-lived, rarely changes
    FileMetadataTTL     = 30 * time.Minute  // Medium-lived
    APIResponseTTL      = 5 * time.Minute   // Short-lived, for identical requests
)
```

### 7. Invalidation Strategy

```go
// On entity update/delete, invalidate related caches
func (d *CachedUserRepo) Update(ctx context.Context, user *User) error {
    // Update database
    if err := d.repo.Update(ctx, user); err != nil {
        return err
    }
    
    // Invalidate related caches
    d.cache.DeleteByPattern(ctx, fmt.Sprintf("user:%s:*", user.ID()))
    d.cache.Delete(ctx, fmt.Sprintf("users:list:*"))
    
    return nil
}
```

## Use Cases

### 1. User Sessions
```go
// Cache session data to reduce database hits
cachedSessionStore := cache.DecorateSessionStore(sessionStore, cache, 15*time.Minute)
```

### 2. Role Permissions
```go
// Cache role permissions (rarely change)
cachedRoleRepo := cache.DecorateRoleRepo(roleRepo, cache, 2*time.Hour)
```

### 3. File Metadata
```go
// Cache frequently accessed file metadata
cachedFileRepo := cache.DecorateFileRepo(fileRepo, cache, 30*time.Minute)
```

### 4. API Responses
```go
// Cache expensive query results
api.GET("/users", 
    middleware.Cache(cache, 5*time.Minute),
    handler.ListUsers)
```

### 5. Download URLs
```go
// Cache presigned URLs (short TTL)
key := fmt.Sprintf("download:url:%s", fileID)
cache.Set(ctx, key, presignedURL, 5*time.Minute)
```

## Configuration

```yaml
cache:
  type: redis  # redis | memory
  redis:
    url: redis://localhost:6379/1
    prefix: skeleton
    pool_size: 10
  memory:
    max_size: 10000
    cleanup_interval: 1m
  default_ttl: 300s
```

## Implementation Layers

### Layer 1: Core Interface (`pkg/cache/cache.go`)
- Cache interface definition
- Common errors
- Helper functions

### Layer 2: Implementations (`pkg/cache/`)
- `memory.go` - In-memory cache with TTL
- `redis.go` - Redis-backed cache

### Layer 3: Middleware (`pkg/middleware/cache.go`)
- HTTP response caching
- Cache key generation
- Cache invalidation hooks

### Layer 4: Decorators (`pkg/cache/decorator.go`)
- Repository decorators
- Automatic invalidation
- Cache-aside pattern

## Trade-offs

### Positive
- ✅ Significantly reduces database load (up to 90% reduction)
- ✅ Improves response times (50-500ms -> <5ms)
- ✅ Scales horizontally with Redis
- ✅ Reduces infrastructure costs (smaller DB instances)
- ✅ Pluggable architecture (easy to switch implementations)
- ✅ Graceful degradation (cache miss -> database)

### Negative
- ❌ Adds complexity (cache invalidation strategies)
- ❌ Potential data staleness (eventual consistency)
- ❌ Additional infrastructure (Redis for production)
- ❌ Memory usage for in-memory cache
- ❌ Requires careful TTL management

### Neutral
- Cache-aside is deceptively simple but requires careful thought
- Memory cache isonly suitable for single-instance deployments
- Redis cache requires operational knowledge

## Testing Strategy

### Unit Tests
- Mock Cache interface
- Test cache hit/miss scenarios
- Test TTL expiration
- Test pattern-based deletion

### Integration Tests
- Memory cache with concurrent access
- Redis cache with real Redis instance
- Cache invalidation after updates
- Fallback to database on cache miss

### Performance Tests
- Cache hit ratio measurement
- Response time comparisons
- Memory usage monitoring
- Load testing with/without cache

## Migration Plan

### Phase 1: Infrastructure
1. Create `pkg/cache` package
2. Implement Cache interface
3. Implement Memory and Redis caches
4. Add configuration support

### Phase 2: Integration
1. Add cache decorators for repositories
2. Integrate with Identity context
3. Integrate with Files context
4. Add HTTP cache middleware

### Phase 3: Monitoring
1. Add cache metrics (hit ratio, latency)
2. Add logging for cache operations
3. Monitor memory usage
4. Set up alerting for cache failures

## Security Considerations

1. **Sensitive Data**: Never cache passwords, tokens, or secrets
2. **User Isolation**: Include user context in cache keys
3. **TTL Management**: Short TTL for sensitive operations
4. **Encryption**: Consider encrypting cached data in Redis
5. **Access Control**: Redis AUTH and network isolation

## Observability

```go
// Metrics to track
- cache_hit_total{cache_type, context}
- cache_miss_total{cache_type, context}
- cache_latency_seconds{cache_type, operation}
- cache_memory_bytes{cache_type}
- cache_errors_total{cache_type, error_type}
```

## References

- [Redis Caching Best Practices](https://redis.io/docs/manual/client-side-caching/)
- [Cache-Aside Pattern](https://docs.microsoft.com/en-us/azure/architecture/patterns/cache-aside)
- [Go Cache Libraries Comparison](https://github.com/golang/go/wiki/PackageGoTools)
- [ADR-003: Event Bus](ADR-003-event-bus.md) - Redis usage for events
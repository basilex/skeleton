// Package main provides dependency injection wiring for the API server.
// This file contains the manual dependency injection setup that wires together
// all application components following the Dependency Inversion Principle.
//
// The wiring process creates the complete dependency graph including:
//   - Domain services (token service, password hasher)
//   - Repositories (user, role, audit, notification, template, preferences)
//   - Application handlers (commands and queries for each bounded context)
//   - HTTP handlers (ports layer)
//   - Middlewares (session, auth, RBAC)
//   - Event bus and event handlers
//   - Background workers (notification worker)
//
// All dependencies are centrally managed through the Dependencies struct,
// making the application configuration explicit and testable.
package main

import (
	"context"
	"fmt"
	"time"

	auditCommand "github.com/basilex/skeleton/internal/audit/application/command"
	auditQuery "github.com/basilex/skeleton/internal/audit/application/query"
	auditEventHandler "github.com/basilex/skeleton/internal/audit/infrastructure/eventhandler"
	auditPersistence "github.com/basilex/skeleton/internal/audit/infrastructure/persistence"
	auditHTTP "github.com/basilex/skeleton/internal/audit/ports/http"
	filesCommand "github.com/basilex/skeleton/internal/files/application/command"
	filesQuery "github.com/basilex/skeleton/internal/files/application/query"
	filesHandler "github.com/basilex/skeleton/internal/files/infrastructure/handler"
	filesPersistence "github.com/basilex/skeleton/internal/files/infrastructure/persistence"
	filesProcessing "github.com/basilex/skeleton/internal/files/infrastructure/processing"
	filesStorage "github.com/basilex/skeleton/internal/files/infrastructure/storage"
	filesHTTP "github.com/basilex/skeleton/internal/files/ports/http"
	identityCommand "github.com/basilex/skeleton/internal/identity/application/command"
	identityQuery "github.com/basilex/skeleton/internal/identity/application/query"
	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/identity/infrastructure/persistence"
	"github.com/basilex/skeleton/internal/identity/infrastructure/session"
	"github.com/basilex/skeleton/internal/identity/infrastructure/token"
	identityHTTP "github.com/basilex/skeleton/internal/identity/ports/http"
	notificationCommand "github.com/basilex/skeleton/internal/notifications/application/command"
	notificationEventHandler "github.com/basilex/skeleton/internal/notifications/application/eventhandler"
	notificationQuery "github.com/basilex/skeleton/internal/notifications/application/query"
	notificationPersistence "github.com/basilex/skeleton/internal/notifications/infrastructure/persistence"
	"github.com/basilex/skeleton/internal/notifications/infrastructure/sender"
	"github.com/basilex/skeleton/internal/notifications/infrastructure/worker"
	notificationHTTP "github.com/basilex/skeleton/internal/notifications/ports/http"
	"github.com/basilex/skeleton/internal/status/application/query"
	statusDomain "github.com/basilex/skeleton/internal/status/domain"
	statusHTTP "github.com/basilex/skeleton/internal/status/ports/http"
	tasksDomain "github.com/basilex/skeleton/internal/tasks/domain"
	"github.com/basilex/skeleton/pkg/cache"
	"github.com/basilex/skeleton/pkg/config"
	"github.com/basilex/skeleton/pkg/eventbus"
	membus "github.com/basilex/skeleton/pkg/eventbus/memory"
	redisbus "github.com/basilex/skeleton/pkg/eventbus/redis"
	"github.com/basilex/skeleton/pkg/ratelimit"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// Cache defines the interface for caching operations.
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)
	SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error
	DeleteMulti(ctx context.Context, keys []string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Clear(ctx context.Context) error
}

// RateLimiter defines the interface for rate limiting operations.
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
	AllowN(ctx context.Context, key string, n int) (bool, error)
	Remaining(ctx context.Context, key string) (int, error)
	Reset(ctx context.Context, key string) error
	Configure(config ratelimit.Config) error
}

// Dependencies holds all initialized application components required to run the API server.
// This struct serves as a container for the complete dependency graph, making dependencies
// explicit and providing easy access throughout the application lifecycle.
//
// Dependencies are organized by domain context:
//   - Infrastructure: EventBus for domain event propagation
//   - Identity: Authentication, authorization, and user management
//   - Status: System health and build information
//   - Audit: Audit log recording and querying
//   - Notifications: Notification management and background processing
type Dependencies struct {
	// EventBus provides the domain event bus for publishing and subscribing to events.
	EventBus eventbus.Bus

	// Infrastructure components
	Config      *config.Config // Application configuration
	Database    *pgxpool.Pool  // PostgreSQL connection pool
	RedisClient *redis.Client  // Redis client (optional)

	// Cache provides distributed caching functionality.
	Cache Cache
	// RateLimiter provides rate limiting for API protection.
	RateLimiter RateLimiter

	// SessionMiddleware handles cookie-based session authentication.
	SessionMiddleware *session.Middleware
	// AuthMiddleware handles JWT token validation and authentication.
	AuthMiddleware *identityHTTP.AuthMiddleware
	// RBACMiddleware enforces role-based access control for protected endpoints.
	RBACMiddleware *identityHTTP.RBACMiddleware

	// IdentityHandler handles all HTTP endpoints related to user authentication and management.
	IdentityHandler *identityHTTP.Handler
	// StatusHandler handles health check and build information endpoints.
	StatusHandler *statusHTTP.Handler
	// AuditHandler handles audit log query endpoints.
	AuditHandler *auditHTTP.Handler
	// AuditEventHandler processes identity-related events for audit logging.
	AuditEventHandler *auditEventHandler.IdentityEventHandler
	// NotificationHandler handles all notification-related HTTP endpoints.
	NotificationHandler *notificationHTTP.Handler
	// NotificationWorker processes notification sending in the background.
	NotificationWorker *worker.NotificationWorker
	// NotificationEventHandler processes identity events for automatic notifications.
	NotificationEventHandler *notificationEventHandler.IdentityEventHandler
	// FilesHandler handles all file-related HTTP endpoints.
	FilesHandler *filesHTTP.Handler
	// ProcessFileHandler handles file processing tasks.
	ProcessFileHandler *filesHandler.ProcessFileHandler
}

// wireDependencies constructs the complete dependency graph for the application.
// It creates and wires together all repositories, services, handlers, and middlewares
// following the dependency inversion principle.
//
// The wiring process follows this hierarchy:
//  1. Infrastructure: Event bus for domain events
//  2. Repositories: Data persistence adapters for each bounded context
//  3. Domain services: Token service, password hasher
//  4. Application layer: Command and query handlers for each use case
//  5. HTTP layer: Request handlers and middlewares
//  6. Event handlers: Domain event subscribers
//  7. Background workers: Notification worker
//
// Parameters:
//   - cfg: Application configuration loaded from environment/file
//   - db: Database connection for SQLite
//   - version: Application version for status endpoint
//   - commit: Git commit hash for status endpoint
//   - buildTime: Build timestamp for status endpoint
//   - goVersion: Go runtime version for status endpoint
//
// Returns a fully initialized Dependencies struct containing all application components.
func wireDependencies(cfg *config.Config, pool *pgxpool.Pool, redisClient *redis.Client, version, commit, buildTime, goVersion string) *Dependencies {
	// Initialize event bus.
	// Uses Redis Pub/Sub in production, in-memory for development.
	var bus eventbus.Bus
	if redisClient != nil {
		bus = redisbus.New(redisClient)
	} else {
		bus = membus.New()
	}

	// Initialize caching infrastructure.
	// Uses Redis cache in production, in-memory for development.
	var cacheClient Cache
	if redisClient != nil {
		cacheClient = cache.NewRedisCache(redisClient, cfg.Cache.RedisPrefix)
	} else {
		cleanupInterval := time.Duration(cfg.Cache.CleanupInterval) * time.Second
		if cleanupInterval == 0 {
			cleanupInterval = time.Minute
		}
		cacheClient = cache.NewMemoryCache(cleanupInterval)
	}

	// Initialize rate limiting infrastructure.
	// Uses Redis sliding window in production, in-memory token bucket for development.
	var rateLimiter RateLimiter
	if redisClient != nil {
		config := ratelimit.Config{
			Limit:     cfg.RateLimit.Global.Limit,
			Window:    time.Duration(cfg.RateLimit.Global.Window) * time.Second,
			KeyPrefix: cfg.RateLimit.KeyPrefix,
		}
		rateLimiter = ratelimit.NewSlidingWindow(redisClient, config)
	} else {
		config := ratelimit.Config{
			Limit:     cfg.RateLimit.Global.Limit,
			Window:    time.Duration(cfg.RateLimit.Global.Window) * time.Second,
			KeyPrefix: cfg.RateLimit.KeyPrefix,
		}
		rateLimiter = ratelimit.NewTokenBucket(config)
	}

	// Identity Bounded Context
	// Initialize repositories for the identity context.
	userRepo := persistence.NewUserRepository(pool)
	roleRepo := persistence.NewRoleRepository(pool)
	// Initialize audit repository (shared across contexts for audit logging).
	auditRepo := auditPersistence.NewAuditRepository(pool)

	// Initialize domain services for authentication and password hashing.
	tokenService := newTokenService(cfg)
	passwordHasher := &domain.BcryptHasher{}

	// Initialize session management (cookie-based authentication).
	sessionStore := newSessionStore(cfg)
	sessionMiddleware := session.NewMiddleware(sessionStore, cfg.Session)

	// Initialize command handlers for identity use cases.
	registerHandler := identityCommand.NewRegisterUserHandler(userRepo, roleRepo, bus, passwordHasher)
	loginHandler := identityCommand.NewLoginUserHandler(userRepo, roleRepo, tokenService, bus)
	logoutHandler := identityCommand.NewLogoutUserHandler(userRepo, bus)
	assignRoleHandler := identityCommand.NewAssignRoleHandler(userRepo, roleRepo, bus)
	revokeRoleHandler := identityCommand.NewRevokeRoleHandler(userRepo, roleRepo, bus)
	// Initialize query handlers for identity use cases.
	getUserHandler := identityQuery.NewGetUserHandler(userRepo, roleRepo)
	listUsersHandler := identityQuery.NewListUsersHandler(userRepo, roleRepo)

	// Initialize HTTP handler for identity endpoints.
	identityHandler := identityHTTP.NewHandler(
		registerHandler,
		loginHandler,
		logoutHandler,
		assignRoleHandler,
		revokeRoleHandler,
		getUserHandler,
		listUsersHandler,
		sessionStore,
	)

	// Initialize authentication and RBAC middlewares for endpoint protection.

	authMiddleware := identityHTTP.NewAuthMiddleware(tokenService)
	rbacMiddleware := identityHTTP.NewRBACMiddleware()

	// Status Bounded Context
	statusHandler := newStatusHandler(version, commit, buildTime, goVersion, cfg.App.Env)

	// Audit Bounded Context
	// Initialize audit command and query handlers.
	logEventHandler := auditCommand.NewLogEventHandler(auditRepo)
	listRecordsHandler := auditQuery.NewListRecordsHandler(auditRepo)
	auditHandler := auditHTTP.NewHandler(listRecordsHandler)
	// Initialize event handler for recording identity events in audit log.
	auditIdentityEventHandler := auditEventHandler.NewIdentityEventHandler(logEventHandler)

	// Notifications Bounded Context
	// Initialize repositories for notification persistence.
	notificationRepo := notificationPersistence.NewNotificationRepository(pool)
	templateRepo := notificationPersistence.NewTemplateRepository(pool)
	preferencesRepo := notificationPersistence.NewPreferencesRepository(pool)

	// Initialize command handlers for notification use cases.
	createNotificationHandler := notificationCommand.NewCreateNotificationHandler(notificationRepo, bus)
	createFromTemplateHandler := notificationCommand.NewCreateFromTemplateHandler(notificationRepo, templateRepo)
	markDeliveredHandler := notificationCommand.NewMarkDeliveredHandler(notificationRepo)
	markFailedHandler := notificationCommand.NewMarkFailedHandler(notificationRepo)
	updatePreferencesHandler := notificationCommand.NewUpdatePreferencesHandler(preferencesRepo)
	// Initialize query handlers for notification use cases.
	getNotificationHandler := notificationQuery.NewGetNotificationHandler(notificationRepo)
	listNotificationsHandler := notificationQuery.NewListNotificationsHandler(notificationRepo)
	getPreferencesHandler := notificationQuery.NewGetPreferencesHandler(preferencesRepo)
	// Initialize command and query handlers for template management.
	createTemplateHandler := notificationCommand.NewCreateTemplateHandler(templateRepo)
	updateTemplateHandler := notificationCommand.NewUpdateTemplateHandler(templateRepo)
	getTemplateHandler := notificationQuery.NewGetTemplateHandler(templateRepo)
	listTemplatesHandler := notificationQuery.NewListTemplatesHandler(templateRepo)

	// Initialize HTTP handler for notification endpoints.
	notificationHandler := notificationHTTP.NewHandler(
		createNotificationHandler,
		createFromTemplateHandler,
		markDeliveredHandler,
		markFailedHandler,
		updatePreferencesHandler,
		getNotificationHandler,
		listNotificationsHandler,
		getPreferencesHandler,
		createTemplateHandler,
		updateTemplateHandler,
		getTemplateHandler,
		listTemplatesHandler,
	)

	// Initialize notification sender infrastructure.
	// Notification sender (console for development - logs to console)
	// In production, replace with real SMTP/AWS SES/Twilio implementations
	emailSender := sender.NewConsoleEmailSender()
	compositeSender := sender.NewCompositeSender(emailSender, nil, nil, nil)

	// Initialize the notification background worker for async processing.
	notificationWorker := worker.NewNotificationWorker(
		notificationRepo,
		preferencesRepo,
		compositeSender,
		bus,
		worker.WorkerConfig{
			PollInterval:   5e9, // 5 seconds
			BatchSize:      100,
			StalledTimeout: 300e9, // 5 minutes
		},
	)

	// Initialize event handler for automatic notifications on identity events.
	notificationIdentityEventHandler := notificationEventHandler.NewIdentityEventHandler(createFromTemplateHandler)

	// Files bounded context
	fileRepo := filesPersistence.NewFileRepository(pool)
	uploadRepo := filesPersistence.NewUploadRepository(pool)
	processingRepo := filesPersistence.NewProcessingRepository(pool)

	localStorage, err := filesStorage.NewLocalStorage("./uploads", "http://localhost:8080/uploads")
	if err != nil {
		panic(fmt.Errorf("create local storage: %w", err))
	}

	imageProcessor := filesProcessing.NewImagingProcessor()

	// Files command handlers
	uploadFileHandler := filesCommand.NewUploadFileHandler(fileRepo, localStorage, bus)
	deleteFileHandler := filesCommand.NewDeleteFileHandler(fileRepo, localStorage, bus)
	requestUploadURLHandler := filesCommand.NewRequestUploadURLHandler(uploadRepo, fileRepo, localStorage)
	confirmUploadHandler := filesCommand.NewConfirmUploadHandler(uploadRepo, fileRepo, bus)

	// Note: ProcessFileHandler uses Tasks context which is not yet integrated
	// For now, passing nil for taskRepo - this needs to be added when Tasks is integrated
	var taskRepo tasksDomain.TaskRepository // Placeholder - Tasks integration needed
	_ = taskRepo                            // Avoid unused variable error until Tasks is integrated
	requestProcessingHandler := filesCommand.NewRequestProcessingHandler(processingRepo, fileRepo, taskRepo)

	// Files query handlers
	getFileHandler := filesQuery.NewGetFileHandler(fileRepo)
	listFilesHandler := filesQuery.NewListFilesHandler(fileRepo)
	getProcessingStatusHandler := filesQuery.NewGetProcessingStatusHandler(processingRepo)
	listProcessingsHandler := filesQuery.NewListProcessingsHandler(processingRepo)

	// Files HTTP handler
	filesHTTPHandler := filesHTTP.NewHandler(
		uploadFileHandler,
		deleteFileHandler,
		requestUploadURLHandler,
		confirmUploadHandler,
		requestProcessingHandler,
		getFileHandler,
		listFilesHandler,
		getProcessingStatusHandler,
		listProcessingsHandler,
	)

	// ProcessFileHandler for Tasks integration
	processFileHandler := filesHandler.NewProcessFileHandler(
		processingRepo,
		fileRepo,
		localStorage,
		imageProcessor,
	)

	// Register domain event handlers with the event bus.
	// Register audit event handler to record identity events.
	auditIdentityEventHandler.Register(bus)
	// Register notification event handler to send automatic notifications.
	notificationIdentityEventHandler.Register(bus)

	// Return the complete dependency graph.
	return &Dependencies{
		EventBus:                 bus,
		Config:                   cfg,
		Database:                 pool,
		RedisClient:              redisClient,
		Cache:                    cacheClient,
		RateLimiter:              rateLimiter,
		IdentityHandler:          identityHandler,
		AuthMiddleware:           authMiddleware,
		RBACMiddleware:           rbacMiddleware,
		SessionMiddleware:        sessionMiddleware,
		StatusHandler:            statusHandler,
		AuditHandler:             auditHandler,
		AuditEventHandler:        auditIdentityEventHandler,
		NotificationHandler:      notificationHandler,
		NotificationWorker:       notificationWorker,
		NotificationEventHandler: notificationIdentityEventHandler,
		FilesHandler:             filesHTTPHandler,
		ProcessFileHandler:       processFileHandler,
	}
}

// newSessionStore creates a session store for managing user sessions.
// Currently uses in-memory storage suitable for development and single-instance deployments.
// For distributed deployments, consider using Redis-backed session storage.
//
// Parameters:
//   - cfg: Application configuration containing session TTL settings
//
// Returns an in-memory session store with the configured TTL.
func newSessionStore(cfg *config.Config) session.Store {
	return session.NewInMemoryStore(cfg.Session.TTLMinutes)
}

// newTokenService creates the token service for JWT generation and validation.
// In development and test environments, uses a mock token service for simplified testing.
// In production environments, uses RSA-based JWT tokens loaded from configured key files.
//
// The token service supports:
//   - Access token generation and validation
//   - Refresh token generation
//   - Public key verification for token validation
//
// Parameters:
//   - cfg: Application configuration containing auth key paths and TTL settings
//
// Returns a TokenService implementation appropriate for the environment.
// Panics if JWT key loading fails in production mode.
func newTokenService(cfg *config.Config) domain.TokenService {
	if cfg.App.Env == "dev" || cfg.App.Env == "test" {
		return &token.MockTokenService{}
	}

	ts, err := token.NewJWTService(
		cfg.Auth.PrivateKeyPath,
		cfg.Auth.PublicKeyPath,
		cfg.Auth.AccessTTLMinutes,
		cfg.Auth.RefreshTTLDays,
	)
	if err != nil {
		panic("failed to create JWT service: " + err.Error())
	}
	return ts
}

// newStatusHandler creates the HTTP handler for status and health check endpoints.
// Initializes build information with version, commit, build time, and Go version
// for the status endpoint response.
//
// Parameters:
//   - version: Application version string
//   - commit: Git commit hash
//   - buildTime: Build timestamp
//   - goVersion: Go runtime version
//   - env: Application environment (dev, test, prod)
//
// Returns a status HTTP handler configured with build information.
// Panics if build info creation fails (which should never occur with valid inputs).
func newStatusHandler(version, commit, buildTime, goVersion, env string) *statusHTTP.Handler {
	buildInfo, err := statusDomain.NewBuildInfo(version, commit, buildTime, goVersion, env)
	if err != nil {
		panic("failed to create build info: " + err.Error())
	}
	getBuildInfoHandler := query.NewGetBuildInfoHandler(buildInfo)
	return statusHTTP.NewHandler(getBuildInfoHandler)
}

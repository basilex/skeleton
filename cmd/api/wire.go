package main

import (
	auditCommand "github.com/basilex/skeleton/internal/audit/application/command"
	auditQuery "github.com/basilex/skeleton/internal/audit/application/query"
	auditEventHandler "github.com/basilex/skeleton/internal/audit/infrastructure/eventhandler"
	auditPersistence "github.com/basilex/skeleton/internal/audit/infrastructure/persistence"
	auditHTTP "github.com/basilex/skeleton/internal/audit/ports/http"
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
	"github.com/basilex/skeleton/pkg/config"
	"github.com/basilex/skeleton/pkg/eventbus"
	membus "github.com/basilex/skeleton/pkg/eventbus/memory"
	"github.com/jmoiron/sqlx"
)

type Dependencies struct {
	EventBus eventbus.Bus

	SessionMiddleware *session.Middleware
	AuthMiddleware    *identityHTTP.AuthMiddleware
	RBACMiddleware    *identityHTTP.RBACMiddleware

	IdentityHandler          *identityHTTP.Handler
	StatusHandler            *statusHTTP.Handler
	AuditHandler             *auditHTTP.Handler
	AuditEventHandler        *auditEventHandler.IdentityEventHandler
	NotificationHandler      *notificationHTTP.Handler
	NotificationWorker       *worker.NotificationWorker
	NotificationEventHandler *notificationEventHandler.IdentityEventHandler
}

func wireDependencies(cfg *config.Config, db *sqlx.DB, version, commit, buildTime, goVersion string) *Dependencies {
	bus := membus.New()

	// Identity
	userRepo := persistence.NewUserRepository(db)
	roleRepo := persistence.NewRoleRepository(db)
	auditRepo := auditPersistence.NewAuditRepository(db)

	tokenService := newTokenService(cfg)
	passwordHasher := &domain.BcryptHasher{}

	sessionStore := newSessionStore(cfg)
	sessionMiddleware := session.NewMiddleware(sessionStore, cfg.Session)

	registerHandler := identityCommand.NewRegisterUserHandler(userRepo, roleRepo, bus, passwordHasher)
	loginHandler := identityCommand.NewLoginUserHandler(userRepo, roleRepo, tokenService, bus)
	logoutHandler := identityCommand.NewLogoutUserHandler(userRepo, bus)
	assignRoleHandler := identityCommand.NewAssignRoleHandler(userRepo, roleRepo, bus)
	revokeRoleHandler := identityCommand.NewRevokeRoleHandler(userRepo, roleRepo, bus)
	getUserHandler := identityQuery.NewGetUserHandler(userRepo, roleRepo)
	listUsersHandler := identityQuery.NewListUsersHandler(userRepo, roleRepo)

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

	authMiddleware := identityHTTP.NewAuthMiddleware(tokenService)
	rbacMiddleware := identityHTTP.NewRBACMiddleware()

	// Status
	statusHandler := newStatusHandler(version, commit, buildTime, goVersion, cfg.App.Env)

	// Audit
	logEventHandler := auditCommand.NewLogEventHandler(auditRepo)
	listRecordsHandler := auditQuery.NewListRecordsHandler(auditRepo)
	auditHandler := auditHTTP.NewHandler(listRecordsHandler)
	auditIdentityEventHandler := auditEventHandler.NewIdentityEventHandler(logEventHandler)

	// Notifications
	notificationRepo := notificationPersistence.NewNotificationRepository(db)
	templateRepo := notificationPersistence.NewTemplateRepository(db)
	preferencesRepo := notificationPersistence.NewPreferencesRepository(db)

	createNotificationHandler := notificationCommand.NewCreateNotificationHandler(notificationRepo, bus)
	createFromTemplateHandler := notificationCommand.NewCreateFromTemplateHandler(notificationRepo, templateRepo)
	markDeliveredHandler := notificationCommand.NewMarkDeliveredHandler(notificationRepo)
	markFailedHandler := notificationCommand.NewMarkFailedHandler(notificationRepo)
	updatePreferencesHandler := notificationCommand.NewUpdatePreferencesHandler(preferencesRepo)
	getNotificationHandler := notificationQuery.NewGetNotificationHandler(notificationRepo)
	listNotificationsHandler := notificationQuery.NewListNotificationsHandler(notificationRepo)
	getPreferencesHandler := notificationQuery.NewGetPreferencesHandler(preferencesRepo)
	createTemplateHandler := notificationCommand.NewCreateTemplateHandler(templateRepo)
	updateTemplateHandler := notificationCommand.NewUpdateTemplateHandler(templateRepo)
	getTemplateHandler := notificationQuery.NewGetTemplateHandler(templateRepo)
	listTemplatesHandler := notificationQuery.NewListTemplatesHandler(templateRepo)

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

	// Notification sender (console for development - logs to console)
	// In production, replace with real SMTP/AWS SES/Twilio implementations
	emailSender := sender.NewConsoleEmailSender()
	compositeSender := sender.NewCompositeSender(emailSender, nil, nil, nil)

	// Notification worker
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

	// Notification event handler (auto-send emails on user registration)
	notificationIdentityEventHandler := notificationEventHandler.NewIdentityEventHandler(createFromTemplateHandler)

	// Register event handlers
	auditIdentityEventHandler.Register(bus)
	notificationIdentityEventHandler.Register(bus)

	return &Dependencies{
		IdentityHandler:          identityHandler,
		AuthMiddleware:           authMiddleware,
		RBACMiddleware:           rbacMiddleware,
		SessionMiddleware:        sessionMiddleware,
		StatusHandler:            statusHandler,
		EventBus:                 bus,
		AuditHandler:             auditHandler,
		AuditEventHandler:        auditIdentityEventHandler,
		NotificationHandler:      notificationHandler,
		NotificationWorker:       notificationWorker,
		NotificationEventHandler: notificationIdentityEventHandler,
	}
}

func newSessionStore(cfg *config.Config) session.Store {
	return session.NewInMemoryStore(cfg.Session.TTLMinutes)
}

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

func newStatusHandler(version, commit, buildTime, goVersion, env string) *statusHTTP.Handler {
	buildInfo, err := statusDomain.NewBuildInfo(version, commit, buildTime, goVersion, env)
	if err != nil {
		panic("failed to create build info: " + err.Error())
	}
	getBuildInfoHandler := query.NewGetBuildInfoHandler(buildInfo)
	return statusHTTP.NewHandler(getBuildInfoHandler)
}

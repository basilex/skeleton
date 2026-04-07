package main

import (
	"github.com/basilex/skeleton/internal/identity/application/command"
	identityQuery "github.com/basilex/skeleton/internal/identity/application/query"
	"github.com/basilex/skeleton/internal/identity/domain"
	"github.com/basilex/skeleton/internal/identity/infrastructure/persistence"
	"github.com/basilex/skeleton/internal/identity/infrastructure/session"
	"github.com/basilex/skeleton/internal/identity/infrastructure/token"
	identityHTTP "github.com/basilex/skeleton/internal/identity/ports/http"
	"github.com/basilex/skeleton/internal/status/application/query"
	statusDomain "github.com/basilex/skeleton/internal/status/domain"
	statusHTTP "github.com/basilex/skeleton/internal/status/ports/http"
	"github.com/basilex/skeleton/pkg/config"
	"github.com/basilex/skeleton/pkg/eventbus"
	membus "github.com/basilex/skeleton/pkg/eventbus/memory"
	"github.com/jmoiron/sqlx"
)

type Dependencies struct {
	IdentityHandler   *identityHTTP.Handler
	AuthMiddleware    *identityHTTP.AuthMiddleware
	RBACMiddleware    *identityHTTP.RBACMiddleware
	SessionMiddleware *session.Middleware
	StatusHandler     *statusHTTP.Handler
	EventBus          eventbus.Bus
}

func wireDependencies(cfg *config.Config, db *sqlx.DB, version, commit, buildTime, goVersion string) *Dependencies {
	bus := membus.New()

	userRepo := persistence.NewUserRepository(db)
	roleRepo := persistence.NewRoleRepository(db)

	tokenService := newTokenService(cfg)
	passwordHasher := &domain.BcryptHasher{}

	sessionStore := newSessionStore(cfg)
	sessionMiddleware := session.NewMiddleware(sessionStore, cfg.Session)

	registerHandler := command.NewRegisterUserHandler(userRepo, roleRepo, bus, passwordHasher)
	loginHandler := command.NewLoginUserHandler(userRepo, roleRepo, tokenService)
	assignRoleHandler := command.NewAssignRoleHandler(userRepo, roleRepo, bus)
	revokeRoleHandler := command.NewRevokeRoleHandler(userRepo, roleRepo, bus)
	getUserHandler := identityQuery.NewGetUserHandler(userRepo, roleRepo)
	listUsersHandler := identityQuery.NewListUsersHandler(userRepo, roleRepo)

	identityHandler := identityHTTP.NewHandler(
		registerHandler,
		loginHandler,
		assignRoleHandler,
		revokeRoleHandler,
		getUserHandler,
		listUsersHandler,
		sessionStore,
	)

	authMiddleware := identityHTTP.NewAuthMiddleware(tokenService)
	rbacMiddleware := identityHTTP.NewRBACMiddleware()

	statusHandler := newStatusHandler(version, commit, buildTime, goVersion, cfg.App.Env)

	return &Dependencies{
		IdentityHandler:   identityHandler,
		AuthMiddleware:    authMiddleware,
		RBACMiddleware:    rbacMiddleware,
		SessionMiddleware: sessionMiddleware,
		StatusHandler:     statusHandler,
		EventBus:          bus,
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

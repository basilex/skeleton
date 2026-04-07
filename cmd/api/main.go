// @title Skeleton API
// @version 1.0
// @description Go DDD Hexagonal architecture skeleton project API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey SessionAuth
// @in cookie
// @name session

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token (format: "Bearer {token}")

// @tag.name auth
// @tag.description Authentication operations
// @tag.name users
// @tag.description User management operations
// @tag.name roles
// @tag.description Role management operations
// @tag.name status
// @tag.description System status and health checks

package main

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/basilex/skeleton/pkg/config"
	"github.com/basilex/skeleton/pkg/database"
	"github.com/basilex/skeleton/pkg/httpserver"
)

var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	setupLogger(cfg)

	if cfg.Database.Path != ":memory:" {
		if err := os.MkdirAll("./data", 0755); err != nil {
			slog.Error("failed to create data directory", "error", err)
			os.Exit(1)
		}
	}

	db, err := database.NewSQLite(cfg.Database.Path)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	di := wireDependencies(cfg, db, version, commit, buildTime, runtime.Version())

	router := setupRouter(cfg, di)

	srv := httpserver.New(router, cfg.App.Port)
	if err := srv.Start(); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}

	<-httpserver.WaitForShutdown(context.Background())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}

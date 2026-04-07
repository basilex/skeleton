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

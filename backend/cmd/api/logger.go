// Package main provides logging configuration for the API server.
// This file contains the logger initialization logic that sets up
// structured logging with configurable level and format.
//
// The logger supports two output formats:
//   - JSON: Structured JSON output (recommended for production)
//   - Text: Human-readable text output (recommended for development)
//
// Log levels can be configured via the LOG_LEVEL environment variable:
//   - debug: Most verbose, includes all log messages
//   - info: Standard operational messages (default)
//   - warn: Warning messages for non-critical issues
//   - error: Error messages for critical issues
package main

import (
	"log/slog"
	"os"

	"github.com/basilex/skeleton/pkg/config"
)

// setupLogger initializes the global structured logger based on configuration.
// Configures the logger with the appropriate level and output format.
//
// Log Level Configuration:
//   - debug: Sets logger to DebugLevel (most verbose)
//   - warn: Sets logger to WarnLevel
//   - error: Sets logger to ErrorLevel (least verbose)
//   - info: Sets logger to InfoLevel (default)
//
// Output Format Configuration:
//   - json: Uses slog.NewJSONHandler for structured JSON output
//   - text: Uses slog.NewTextHandler for human-readable output (default)
//
// The configured logger becomes the global default via slog.SetDefault,
// making it available throughout the application via slog.Info, slog.Debug, etc.
//
// Parameters:
//   - cfg: Application configuration containing log level and format settings
func setupLogger(cfg *config.Config) {
	level := slog.LevelInfo
	switch cfg.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if cfg.Log.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
}

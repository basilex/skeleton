// Package httpserver provides HTTP server utilities with graceful shutdown support.
// It wraps the Gin router with lifecycle management and signal handling.
package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// Server wraps a Gin router with HTTP server lifecycle management.
// It provides graceful startup and shutdown capabilities.
type Server struct {
	router *gin.Engine
	port   string
	srv    *http.Server
}

// New creates a new HTTP server with the given Gin router and port.
// The server is not started until Start is called.
func New(router *gin.Engine, port string) *Server {
	return &Server{
		router: router,
		port:   port,
	}
}

// Start launches the HTTP server in a background goroutine.
// It returns immediately and logs if the server fails after startup.
func (s *Server) Start() error {
	s.srv = &http.Server{
		Addr:    ":" + s.port,
		Handler: s.router,
	}

	go func() {
		slog.Info("starting HTTP server", "port", s.port)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	return nil
}

// Shutdown gracefully stops the HTTP server with a 10-second timeout.
// It waits for active connections to complete before returning.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}
	slog.Info("shutting down HTTP server")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	return nil
}

// WaitForShutdown blocks until an interrupt or termination signal is received.
// It returns a channel that is closed when shutdown should begin.
func WaitForShutdown(ctx context.Context) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		close(done)
	}()
	return done
}

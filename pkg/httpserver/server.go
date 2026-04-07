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

type Server struct {
	router *gin.Engine
	port   string
	srv    *http.Server
}

func New(router *gin.Engine, port string) *Server {
	return &Server{
		router: router,
		port:   port,
	}
}

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

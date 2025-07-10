package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tidy/internal/config"
	"tidy/internal/router"

	"github.com/sirupsen/logrus"
)

type Server struct {
	cfg    *config.Config
	logger *logrus.Logger
}

func New(cfg *config.Config, logger *logrus.Logger) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Server) Run() error {
	handler, err := router.Setup(s.cfg, s.logger)
	if err != nil {
		s.logger.WithError(err).Error("Failed to setup router")
		return fmt.Errorf("failed to setup router: %w", err)
	}

	server := &http.Server{
		Addr:           ":" + s.cfg.Server.Port,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	return s.runWithGracefulShutdown(server)
}

func (s *Server) runWithGracefulShutdown(server *http.Server) error {
	serverErrors := make(chan error, 1)

	go func() {
		s.logger.WithField("addr", server.Addr).Info("Starting HTTP server")
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server failed: %w", err)
		}
		s.logger.Info("Server stopped")
		return nil

	case sig := <-shutdown:
		s.logger.WithField("signal", sig.String()).Info("Shutdown signal received")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		s.logger.Info("Shutting down server...")
		if err := server.Shutdown(ctx); err != nil {
			if closeErr := server.Close(); closeErr != nil {
				s.logger.WithError(closeErr).Error("Failed to force close server")
			}
			return fmt.Errorf("server shutdown failed: %w", err)
		}

		s.logger.Info("Server exited gracefully")
		return nil
	}
}

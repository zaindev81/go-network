package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"tidy/internal/config"
	"tidy/internal/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, logger, err := initializeApp()
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	gin.SetMode(cfg.Server.Mode)

	router := gin.Default()

	router.Use(middleware.Logger(logger))
	router.Use(gin.Recovery())

	corsConfig := createCORSConfig()
	router.Use(cors.New(corsConfig))

	server := &http.Server{
		Addr:           ":" + cfg.Server.Port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	return runServerWithGracefulShutdown(server, logger)
}

func initializeApp() (*config.Config, *logrus.Logger, error) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		return nil, nil, fmt.Errorf("config load failed: %w", err)
	}

	logger := config.InitLogger(cfg)

	logger.WithFields(logrus.Fields{
		"app":     cfg.App.Name,
		"version": cfg.App.Version,
		"port":    cfg.Server.Port,
		"mode":    cfg.Server.Mode,
	}).Info("Application initialized successfully")

	return cfg, logger, nil
}

func createCORSConfig() cors.Config {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
		"X-Requested-With",
		"Accept",
		"Cache-Control",
	}
	config.AllowMethods = []string{
		"GET",
		"POST",
		"PUT",
		"DELETE",
		"OPTIONS",
		"PATCH",
	}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	return config
}

func runServerWithGracefulShutdown(server *http.Server, logger *logrus.Logger) error {
	serverErrors := make(chan error, 1)

	go func() {
		logger.WithField("addr", server.Addr).Info("Starting HTTP server")
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server failed: %w", err)
		}
		logger.Info("Server stopped")
		return nil

	case sig := <-shutdown:
		logger.WithField("signal", sig.String()).Info("Shutdown signal received")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		logger.Info("Shutting down server...")
		if err := server.Shutdown(ctx); err != nil {
			if closeErr := server.Close(); closeErr != nil {
				logger.WithError(closeErr).Error("Failed to force close server")
			}
			return fmt.Errorf("server shutdown failed: %w", err)
		}

		logger.Info("Server exited gracefully")
		return nil
	}
}

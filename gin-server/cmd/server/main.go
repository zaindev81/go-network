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
	"tidy/internal/routes"
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

	router, err := setupRouter(cfg, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to setup router")
		return fmt.Errorf("failed to setup router: %w", err)
	}

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

func setupRouter(cfg *config.Config, logger *logrus.Logger) (*gin.Engine, error) {
	gin.SetMode(cfg.Server.Mode)

	router := gin.New()

	if err := setupMiddleware(router, logger); err != nil {
		return nil, fmt.Errorf("failed to setup middleware: %w", err)
	}

	routes.Setup(router, cfg, logger)

	return router, nil
}

func setupMiddleware(router *gin.Engine, logger *logrus.Logger) error {
	router.Use(middleware.Logger(logger))

	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			logger.WithFields(logrus.Fields{
				"error":  err,
				"path":   c.Request.URL.Path,
				"method": c.Request.Method,
			}).Error("Panic recovered")
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))

	corsConfig := createCORSConfig()
	router.Use(cors.New(corsConfig))

	return nil
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

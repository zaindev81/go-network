package router

import (
	"fmt"
	"net/http"
	"time"

	"tidy/internal/config"
	"tidy/internal/middleware"
	"tidy/internal/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Setup(cfg *config.Config, logger *logrus.Logger) (*gin.Engine, error) {
	gin.SetMode(cfg.Server.Mode)

	router := gin.New()

	if err := setupMiddleware(router, logger); err != nil {
		return nil, fmt.Errorf("failed to setup middleware: %w", err)
	}

	if err := routes.Setup(router, cfg, logger); err != nil {
		return nil, fmt.Errorf("failed to setup routes: %w", err)
	}

	logger.Info("Router setup completed")
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

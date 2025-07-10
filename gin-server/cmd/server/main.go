package main

import (
	"tidy/internal/config"
	"tidy/internal/middleware"
	"tidy/internal/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	logger := config.InitLogger(cfg)

	logger.Info("Starting server")
	logger.WithFields(map[string]interface{}{
		"app":     cfg.App.Name,
		"version": cfg.App.Version,
	}).Info("Server configuration loaded")

	gin.SetMode(cfg.Server.Mode)

	router := gin.Default()

	router.Use(middleware.Logger(logger))
	router.Use(gin.Recovery())

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	router.Use(cors.New(corsConfig))

	routes.Setup(router, cfg, logger)

	logger.Info("Server starting on port " + cfg.Server.Port)
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		logger.Fatal("Failed to start server: " + err.Error())
	}
}

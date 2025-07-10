package routes

import (
	"tidy/internal/config"
	"tidy/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Setup(r *gin.Engine, cfg *config.Config, logger *logrus.Logger) error {
	h := handlers.New(cfg, logger)

	r.GET("/", h.Home)
	r.GET("/status", h.Status)

	r.NoRoute(h.NotFound)

	return nil
}

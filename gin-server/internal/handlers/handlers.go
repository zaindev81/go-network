package handlers

import (
	"net/http"
	"tidy/internal/config"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Handler struct {
	cfg    *config.Config
	logger *logrus.Logger
}

func New(cfg *config.Config, logger *logrus.Logger) *Handler {
	return &Handler{
		cfg:    cfg,
		logger: logger,
	}
}

func (h *Handler) Home(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Status:  "success",
		Message: "Welcome to " + h.cfg.App.Name + "!",
		Data: gin.H{
			"version": h.cfg.App.Version,
		},
	})
}

func (h *Handler) NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, Response{
		Status:  "error",
		Message: "Resource not found",
	})
}

func (h *Handler) Status(c *gin.Context) {
	h.logger.Info("Status endpoint accessed")

	c.JSON(http.StatusOK, Response{
		Status:  "success",
		Message: "Server is running",
		Data: gin.H{
			"timestamp": time.Now(),
			"uptime":    "running",
			"version":   h.cfg.App.Version,
		},
	})
}

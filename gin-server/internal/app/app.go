package app

import (
	"fmt"

	"tidy/internal/config"
	"tidy/internal/server"

	"github.com/sirupsen/logrus"
)

func Run() error {
	cfg, logger, err := initialize()
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	srv := server.New(cfg, logger)
	return srv.Run()
}

func initialize() (*config.Config, *logrus.Logger, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("config load failed: %w", err)
	}

	logger := config.InitLogger(cfg)
	logger.WithFields(map[string]interface{}{
		"app":     cfg.App.Name,
		"version": cfg.App.Version,
		"port":    cfg.Server.Port,
		"mode":    cfg.Server.Mode,
	}).Info("Application initialized successfully")

	return cfg, logger, nil
}

package main

import (
	"data-service/config"
	"data-service/internal/repository/postgres"
	"data-service/pkg/logger"
	"os"
)

func main() {
	cfg := config.MustLoad()

	log := logger.New(logger.Config{
		Level:  cfg.LogConfig.Level,
		Format: cfg.LogConfig.Format,
	})

	if err := postgres.Migrate(cfg.DatabaseConfig.URL, cfg.DatabaseConfig.MigrationsPath); err != nil {
		log.Error("migrations failed", "path", cfg.DatabaseConfig.MigrationsPath, "error", err)
		os.Exit(1)
	}

	log.Info("migrations applied", "path", cfg.DatabaseConfig.MigrationsPath)
}

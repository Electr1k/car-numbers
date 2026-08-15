package main

import (
	"context"
	"data-service/config"
	"data-service/internal/domain"
	"data-service/internal/job"
	"data-service/internal/provider/autonomera"
	"data-service/internal/repository/postgres"
	"data-service/internal/scheduler"
	"data-service/pkg/logger"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api stopped with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.MustLoad()

	log := logger.New(logger.Config{
		Level:  cfg.LogConfig.Level,
		Format: cfg.LogConfig.Format,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info("starting api", "env", cfg.Env)

	database, err := postgres.New(ctx, cfg.DatabaseConfig)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer database.Close()

	sched := scheduler.New(postgres.NewJobRepository(database), log)

	entries, err := scheduleEntries(cfg.AutoNomeraConfig)
	if err != nil {
		return fmt.Errorf("build schedule: %w", err)
	}

	for _, entry := range entries {
		if err := sched.Add(entry); err != nil {
			return fmt.Errorf("register schedule: %w", err)
		}
	}

	return sched.Run(ctx)
}

func scheduleEntries(cfg config.AutoNomeraConfig) ([]scheduler.Entry, error) {
	active, err := importAutonomeraPayload(autonomera.SectionActive, cfg.ImportDepth)
	if err != nil {
		return nil, err
	}

	archive, err := importAutonomeraPayload(autonomera.SectionArchive, cfg.ImportDepth)
	if err != nil {
		return nil, err
	}

	return []scheduler.Entry{
		{
			Name:      domain.JobNameImportAutonomeraOffers,
			Queue:     domain.JobQueueDefault,
			Spec:      "0 * * * *",
			Payload:   active,
			UniqueKey: job.ImportAutonomeraUniqueKey(autonomera.SectionActive),
		},
		{
			Name:      domain.JobNameImportAutonomeraOffers,
			Queue:     domain.JobQueueDefault,
			Spec:      "30 * * * *",
			Payload:   archive,
			UniqueKey: job.ImportAutonomeraUniqueKey(autonomera.SectionArchive),
		},
	}, nil
}

func importAutonomeraPayload(section autonomera.Section, depth time.Duration) (string, error) {
	encoded, err := json.Marshal(job.ImportAutonomeraPayload{
		Section:   section,
		StopAfter: job.Duration(depth),
	})
	if err != nil {
		return "", fmt.Errorf("marshal %s payload: %w", section, err)
	}

	return string(encoded), nil
}

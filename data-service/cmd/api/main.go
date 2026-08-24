package main

import (
	"context"
	"data-service/config"
	"data-service/internal/job"
	"data-service/internal/job/cron"
	"data-service/internal/repository/postgres"
	"data-service/internal/scheduler"
	"data-service/pkg/logger"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
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

	sched := scheduler.New(log)

	if err := cron.Register(sched, cron.Deps{
		Producer:   job.NewProducer(postgres.NewJobRepository(database)),
		Specs:      cfg.CronConfig,
		AutoNomera: cfg.AutoNomeraConfig,
		Logger:     log,
	}); err != nil {
		return fmt.Errorf("register crons: %w", err)
	}

	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return sched.Run(groupCtx)
	})

	return group.Wait()
}

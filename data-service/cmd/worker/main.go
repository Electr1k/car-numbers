package main

import (
	"context"
	"data-service/config"
	"data-service/internal/domain"
	"data-service/internal/job"
	"data-service/internal/provider/autonomera"
	"data-service/internal/repository/postgres"
	"data-service/internal/usecase"
	"data-service/pkg/logger"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		slog.Error("worker jobs stopped with error", "error", err)
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

	log.Info("starting worker jobs", "env", cfg.Env)

	database, err := postgres.New(ctx, cfg.DatabaseConfig)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer database.Close()

	jobRepository := postgres.NewJobRepository(database)
	offerRepository := postgres.NewOfferRepository(database)

	mapper := autonomera.NewMapper(cfg.AutoNomeraConfig.BaseURL)

	autonomeraClient := autonomera.NewClient(cfg.AutoNomeraConfig.BaseURL, log)

	autonomeraService := autonomera.NewService(
		autonomeraClient,
		mapper,
	)

	importAutonomeraUseCase := usecase.NewImportAutonomeraOffersUseCase(
		autonomeraService,
		offerRepository,
		cfg.AutoNomeraConfig,
		log.With("provider", domain.ProviderAutonomera),
	)

	imortJob, err := job.NewImportAutonomeraJob(jobRepository, *importAutonomeraUseCase)
	resolver := job.NewResolver(jobRepository, imortJob)

	for {
		if ctx.Err() != nil {
			log.Info("worker jobs stopped")
			return nil
		}

		domainJob, err := jobRepository.GetJob(ctx)
		if errors.Is(err, domain.ErrNoJob) {
			time.Sleep(time.Second)
			continue
		}
		if err != nil {
			log.Error("get job", "error", err)
			time.Sleep(time.Second)
			continue
		}
		log.Info("got job", "job", domainJob)

		task, err := resolver.ResolveJob(*domainJob)
		if err != nil {
			err = jobRepository.UpdateJobStatus(ctx, domainJob.Id, domain.JobStatusFailed)
			if err != nil {
				log.Error("failed to update job status", "error", err, "job_id", domainJob.Id)
			}
			time.Sleep(time.Second)
			continue
		}

		err = task.Handle(ctx, domainJob.Payload)
		if err != nil {
			err = jobRepository.UpdateJobStatus(ctx, domainJob.Id, domain.JobStatusFailed)
			if err != nil {
				log.Error("failed to update job status", "error", err, "job_id", domainJob.Id)
			}
			time.Sleep(time.Second)
			continue
		}

		err = jobRepository.UpdateJobStatus(ctx, domainJob.Id, domain.JobStatusSuccess)
		if err != nil {
			log.Error("failed to update job status", "error", err, "job_id", domainJob.Id)
		}

		time.Sleep(time.Second)
	}
}

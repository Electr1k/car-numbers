package main

import (
	"context"
	"data-service/config"
	"data-service/internal/domain"
	"data-service/internal/job"
	"data-service/internal/provider/autonomera"
	"data-service/internal/provider/resolver"
	"data-service/internal/repository/postgres"
	"data-service/internal/usecase"
	"data-service/internal/worker"
	"data-service/pkg/logger"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
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
	featureRepository := postgres.NewFeatureRepository(database)

	autonomeraService := autonomera.NewService(
		autonomera.NewClient(cfg.AutoNomeraConfig.BaseURL, log),
		autonomera.NewMapper(cfg.AutoNomeraConfig.BaseURL),
	)

	importAutonomeraUseCase := usecase.NewImportAutonomeraOffersUseCase(
		autonomeraService,
		offerRepository,
		featureRepository,
		cfg.AutoNomeraConfig,
		log.With("provider", domain.ProviderAutonomera),
	)

	providerResolver := resolver.NewResolver(autonomeraService)
	importOfferDetailUseCase := usecase.NewImportOfferDetailUseCase(
		providerResolver,
		offerRepository,
		featureRepository,
		log,
	)

	jobResolver := job.NewResolver()

	if err := jobResolver.Register(
		domain.JobNameImportAutonomeraOffers,
		job.NewImportAutonomeraJob(jobRepository, importAutonomeraUseCase),
	); err != nil {
		return fmt.Errorf("register jobs: %w", err)
	}

	if err := jobResolver.Register(
		domain.JobNameImportOfferDetailJobName,
		job.NewImportOfferDetailJob(jobRepository, importOfferDetailUseCase),
	); err != nil {
		return fmt.Errorf("register jobs: %w", err)
	}

	return worker.New(jobRepository, jobResolver, domain.JobQueueDefault, log).Run(ctx)
}

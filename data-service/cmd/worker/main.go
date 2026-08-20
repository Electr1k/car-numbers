package main

import (
	"context"
	"data-service/config"
	"data-service/internal/domain"
	"data-service/internal/job"
	"data-service/internal/job/consumer"
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

	// Репозитории
	jobRepository := postgres.NewJobRepository(database)
	offerRepository := postgres.NewOfferRepository(database)
	featureRepository := postgres.NewFeatureRepository(database)

	// Провайдеры
	autonomeraService := autonomera.NewService(
		autonomera.NewClient(cfg.AutoNomeraConfig.BaseURL, log),
		autonomera.NewMapper(cfg.AutoNomeraConfig.BaseURL),
	)
	providerResolver := resolver.NewResolver(autonomeraService)

	// Импорт деталки офферов
	importOfferDetailUseCase := usecase.NewImportOfferDetailUseCase(
		providerResolver,
		offerRepository,
		featureRepository,
		log,
	)
	importOfferDetailJob := job.NewImportOfferDetailJob(jobRepository, importOfferDetailUseCase)

	// Импорт офферов из autonomera
	importAutonomeraUseCase := usecase.NewImportAutonomeraOffersUseCase(
		autonomeraService,
		importOfferDetailJob,
		offerRepository,
		featureRepository,
		cfg.AutoNomeraConfig,
		log.With("provider", domain.ProviderAutonomera),
	)
	importAutonomeraOfferJob := job.NewImportAutonomeraJob(jobRepository, importAutonomeraUseCase)

	jobResolver := consumer.NewResolver()

	if err := jobResolver.Register(
		domain.JobNameImportAutonomeraOffers,
		importAutonomeraOfferJob,
	); err != nil {
		return fmt.Errorf("register jobs: %w", err)
	}

	if err := jobResolver.Register(
		domain.JobNameImportOfferDetail,
		importOfferDetailJob,
	); err != nil {
		return fmt.Errorf("register jobs: %w", err)
	}

	return worker.New(jobRepository, jobResolver, domain.JobQueueDefault, log).Run(ctx)
}

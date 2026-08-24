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

	// Продюсер джоб
	producer := job.NewProducer(jobRepository)

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
	importOfferDetailConsumer := consumer.NewImportOfferDetailConsumer(importOfferDetailUseCase)

	// Импорт офферов из autonomera
	importAutonomeraUseCase := usecase.NewImportAutonomeraOffersUseCase(
		autonomeraService,
		producer,
		offerRepository,
		featureRepository,
		cfg.AutoNomeraConfig,
		log.With("provider", domain.ProviderAutonomera),
	)
	importAutonomeraOffersConsumer := consumer.NewImportAutonomeraOffersConsumer(importAutonomeraUseCase)

	consumerResolver := consumer.NewResolver()

	consumerResolver.Register(
		domain.JobNameImportAutonomeraOffers,
		importAutonomeraOffersConsumer,
	)
	consumerResolver.Register(
		domain.JobNameImportOfferDetail,
		importOfferDetailConsumer,
	)

	return worker.New(jobRepository, consumerResolver, domain.AllJobQueues(), log).Run(ctx)
}

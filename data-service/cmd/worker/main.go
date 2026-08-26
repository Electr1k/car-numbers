package main

import (
	"context"
	"data-service/config"
	"data-service/internal/domain"
	"data-service/internal/feature"
	"data-service/internal/job"
	"data-service/internal/job/consumer"
	"data-service/internal/provider/autonomera"
	"data-service/internal/provider/gosnomeru"
	"data-service/internal/provider/resolver"
	"data-service/internal/repository/postgres"
	"data-service/internal/usecase/importautonomera"
	"data-service/internal/usecase/importgosnomeru"
	"data-service/internal/usecase/importofferdetail"
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

	// Фича-флаги
	featureGuard := feature.NewFeature(featureRepository, log)

	// Продюсер джоб
	producer := job.NewProducer(jobRepository, featureGuard)

	// Провайдеры
	autonomeraService := autonomera.NewService(
		autonomera.NewClient(cfg.AutoNomeraConfig.BaseURL, log),
		autonomera.NewMapper(cfg.AutoNomeraConfig.BaseURL),
	)
	gosnomeruService := gosnomeru.NewService(
		gosnomeru.NewClient(cfg.GosnomeruConfig.BaseURL, log),
		gosnomeru.NewMapper(cfg.GosnomeruConfig.BaseURL),
	)
	providerResolver := resolver.NewResolver(autonomeraService, gosnomeruService)

	// Импорт деталки офферов
	importOfferDetailUseCase := importofferdetail.New(
		providerResolver,
		offerRepository,
		featureGuard,
		log,
	)
	importOfferDetailConsumer := consumer.NewImportOfferDetailConsumer(importOfferDetailUseCase)

	// Импорт офферов из autonomera
	importAutonomeraUseCase := importautonomera.New(
		autonomeraService,
		producer,
		offerRepository,
		featureGuard,
		cfg.AutoNomeraConfig,
		log.With("provider", domain.ProviderAutonomera),
	)
	importAutonomeraOffersConsumer := consumer.NewImportAutonomeraOffersConsumer(importAutonomeraUseCase)

	// Импорт офферов из gosnomeru
	importGosnomeruUseCase := importgosnomeru.New(
		gosnomeruService,
		producer,
		offerRepository,
		featureGuard,
		log.With("provider", domain.ProviderGosnomeru),
	)

	importGosnomeruOffersConsumer := consumer.NewImportGosnomeruOffersConsumer(importGosnomeruUseCase)

	consumerResolver := consumer.NewResolver()

	consumerResolver.Register(
		domain.JobNameImportAutonomeraOffers,
		importAutonomeraOffersConsumer,
	)
	consumerResolver.Register(
		domain.JobNameImportGosnomeruOffers,
		importGosnomeruOffersConsumer,
	)
	consumerResolver.Register(
		domain.JobNameImportOfferDetail,
		importOfferDetailConsumer,
	)

	return worker.New(jobRepository, consumerResolver, domain.AllJobQueues(), log).Run(ctx)
}

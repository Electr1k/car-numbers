package main

import (
	"context"
	"data-service/config"
	"data-service/internal/domain"
	"data-service/internal/providers/autonomera"
	"data-service/internal/repository/postgres"
	"data-service/internal/usecase"
	"data-service/pkg/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	postgreSQL *postgres.Postgres

	autonomeraClient *autonomera.Client

	autonomeraService *autonomera.Service

	offerRepository *postgres.OfferRepository

	importAutonomeraUseCase *usecase.ImportAutonomeraOffersUseCase

	log *slog.Logger
)

func main() {
	cfg := config.MustLoad()

	log = logger.New(logger.Config{
		Level:  cfg.LogConfig.Level,
		Format: cfg.LogConfig.Format,
	})

	log.Info("starting application", "env", cfg.Env)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info("received shutdown signal", "signal", sig)
		cancel()
	}()

	if err := postgres.Migrate(cfg.DatabaseConfig.URL, cfg.DatabaseConfig.MigrationsPath); err != nil {
		log.Error("failed to apply migrations", "error", err)
		os.Exit(1)
	}

	log.Info("migrations applied", "path", cfg.DatabaseConfig.MigrationsPath)

	if err := createRepositories(ctx, cfg); err != nil {
		log.Error("failed to create repositories", "error", err)
		os.Exit(1)
	}
	defer postgreSQL.Close()

	createIntegrationClients(cfg)
	createProviders(cfg)
	createUseCases(cfg)

	log.Info("starting fetch process")

	//Пока догружаем архив, активный листинг поедет по крону
	dataInput := usecase.ImportParams{
		Section:     autonomera.SectionArchive,
		StartOffset: 7000,
		MaxPages:    0,
		StopAfter:   time.Duration(0),
	}
	if err := importAutonomeraUseCase.Handle(ctx, dataInput); err != nil {
		return
	}

	log.Info("fetch process completed successfully")
}

func createRepositories(ctx context.Context, config *config.Config) error {
	var err error
	postgreSQL, err = postgres.New(ctx, config.DatabaseConfig)
	if err != nil {
		return err
	}

	offerRepository = postgres.NewOfferRepository(postgreSQL)

	return nil
}

func createIntegrationClients(cfg *config.Config) {
	autonomeraClient = autonomera.NewClient(cfg.AutoNomeraConfig.BaseURL, log)
}

func createProviders(cfg *config.Config) {
	mapper := autonomera.NewMapper(cfg.AutoNomeraConfig.BaseURL)

	autonomeraService = autonomera.NewService(
		autonomeraClient,
		mapper,
	)
}

func createUseCases(cfg *config.Config) {
	importAutonomeraUseCase = usecase.NewImportAutonomeraOffersUseCase(
		autonomeraService,
		offerRepository,
		cfg.AutoNomeraConfig,
		log.With("provider", domain.ProviderAutonomera),
	)
}

package main

import (
	"context"
	"core/config"
	"core/internal/provider/autonomera"
	"core/internal/repository/postgres"
	usecases "core/internal/usecase/car_number"
	autonomeraIntegration "core/pkg/integration/autonomera"
	"core/pkg/logger"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var (
	postgreSQL *postgres.Postgres

	autonomeraClient *autonomeraIntegration.Client

	autonomeraProvider *autonomera.Provider

	fetchUseCase *usecases.FetchCarNumberUseCase

	log *slog.Logger
)

func main() {
	cfg := config.MustLoad()

	// Initialize logger
	log = logger.New(logger.Config{
		Level:  cfg.LogConfig.Level,
		Format: cfg.LogConfig.Format,
	})

	log.Info("starting application", "env", cfg.Env)

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info("received shutdown signal", "signal", sig)
		cancel()
	}()

	if err := createRepositories(ctx, cfg); err != nil {
		log.Error("failed to create repositories", "error", err)
		os.Exit(1)
	}
	defer postgreSQL.Close()

	createIntegrationClients(cfg)
	createProviders(cfg)
	createUseCases(cfg)

	log.Info("starting fetch process")

	if err := fetchUseCase.Handle(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			log.Info("fetch process cancelled gracefully")
			os.Exit(0)
		}
		log.Error("fetch use case failed", "error", err)
		os.Exit(1)
	}

	log.Info("fetch process completed successfully")
}

func createRepositories(ctx context.Context, cfg *config.Config) error {
	var err error
	postgreSQL, err = postgres.New(ctx, cfg.DatabaseConfig)
	if err != nil {
		return err
	}

	return nil
}

func createIntegrationClients(cfg *config.Config) {
	autonomeraClient = autonomeraIntegration.NewClient(
		autonomeraIntegration.Config{
			BaseURL: cfg.ProviderConfig.AutoNomeraBaseURL,
			Timeout: cfg.ProviderConfig.AutoNomeraTimeout,
		},
		log,
	)
}

func createProviders(cfg *config.Config) {
	parser := autonomera.NewParser()
	autonomeraProvider = autonomera.NewProvider(
		autonomeraClient,
		parser,
		log,
		autonomera.WithBatchSize(cfg.ProviderConfig.BatchSize),
		autonomera.WithStartPosition(cfg.ProviderConfig.StartPosition),
	)
}

func createUseCases(cfg *config.Config) {
	fetchUseCase = usecases.NewFetchCarNumberUseCase(
		autonomeraProvider,
		postgreSQL,
		log,
		usecases.Config{
			StopAfter:      cfg.ProviderConfig.StopAfter,
			RateLimitDelay: cfg.ProviderConfig.RateLimitDelay,
		},
	)
}

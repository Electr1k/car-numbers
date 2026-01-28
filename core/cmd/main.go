package main

import (
	"context"
	"core/config"
	"core/internal/repository/postgres"
	usecases "core/internal/usecase/car_number"
	provides "core/pkg/providers/car_number"
)

var (
	// Repositories
	postgreSQL *postgres.Postgres

	// Services
	autoNomera777 *provides.AutoNomeraHttpClient

	// Usecases
	registerUseCase *usecases.FetchCarNumberUseCase

	//
)

func main() {
	cfg := config.MustLoad()

	println(cfg)
	ctx := context.Background()

	if err := createRepositories(ctx, cfg); err != nil {
		panic(err)
	}

	createServices(cfg)
	createUseCases(cfg)

	defer postgreSQL.Close()

	registerUseCase.Handle()
}

func createRepositories(ctx context.Context, cfg *config.Config) error {
	var err error
	postgreSQL, err = postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}

	return nil
}

func createUseCases(cgf *config.Config) {
	registerUseCase = usecases.NewFetchPricesUseCase(autoNomera777, postgreSQL)
}

func createServices(cgf *config.Config) {
	autoNomera777 = provides.AutoNomeraNewClient()
}

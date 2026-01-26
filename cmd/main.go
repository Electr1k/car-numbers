package main

import (
	"car-numers/config"
	"car-numers/internal/repository/postgres"
	"context"
)

var (
	// Repositories
	postgreSQL *postgres.Postgres
)

func main() {
	cfg := config.MustLoad()

	println(cfg)
	ctx := context.Background()

	if err := createRepositories(ctx, cfg); err != nil {
		panic(err)
	}
	defer postgreSQL.Close()
}

func createRepositories(ctx context.Context, cfg *config.Config) error {
	var err error
	postgreSQL, err = postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}

	return nil
}

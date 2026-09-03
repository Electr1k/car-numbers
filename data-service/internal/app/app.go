package app

import (
	"context"
	"data-service/config"
	"data-service/internal/repository/postgres"
	"data-service/pkg/logger"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// App - общее окружение сервиса
type App struct {
	Config   *config.Config
	Logger   *slog.Logger
	Database *postgres.Postgres
}

func New(ctx context.Context) (*App, error) {
	cfg := config.MustLoad()

	log := logger.New(logger.Config{
		Level:  cfg.LogConfig.Level,
		Format: cfg.LogConfig.Format,
	})

	database, err := postgres.New(ctx, cfg.DatabaseConfig)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return &App{
		Config:   cfg,
		Logger:   log,
		Database: database,
	}, nil
}

func (a *App) Close() {
	a.Database.Close()
}

// Run - запуск сервера замыканием
func Run(name string, fn func(ctx context.Context, app *App) error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, name, fn); err != nil {
		slog.Error(name+" stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, name string, fn func(ctx context.Context, app *App) error) error {
	application, err := New(ctx)
	if err != nil {
		return err
	}
	defer application.Close()

	application.Logger.Info("starting "+name, "env", application.Config.Env)

	return fn(ctx, application)
}

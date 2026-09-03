package main

import (
	"context"
	"data-service/internal/app"
	"data-service/internal/repository/postgres"
	"fmt"
)

func main() {
	app.Run("migrate", func(_ context.Context, a *app.App) error {
		path := a.Config.DatabaseConfig.MigrationsPath

		if err := postgres.Migrate(a.Config.DatabaseConfig.URL, path); err != nil {
			return fmt.Errorf("apply migrations from %s: %w", path, err)
		}

		a.Logger.Info("migrations applied", "path", path)

		return nil
	})
}

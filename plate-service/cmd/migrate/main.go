package main

import (
	"context"
	"fmt"
	"plate-service/internal/app"
	"plate-service/internal/repository/postgres"
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

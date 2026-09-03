package main

import (
	"context"
	"data-service/internal/app"
	"data-service/internal/feature"
	"data-service/internal/job"
	"data-service/internal/job/cron"
	"data-service/internal/repository/postgres"
	"data-service/internal/scheduler"
	"fmt"

	"golang.org/x/sync/errgroup"
)

func main() {
	app.Run("api", func(ctx context.Context, a *app.App) error {
		featureGuard := feature.NewFeature(postgres.NewFeatureRepository(a.Database), a.Logger)

		// Запуск шедулера для кронов
		sched := scheduler.New(a.Logger)
		if err := cron.Register(sched, cron.Deps{
			Producer:   job.NewProducer(postgres.NewJobRepository(a.Database), featureGuard),
			Specs:      a.Config.CronConfig,
			AutoNomera: a.Config.AutoNomeraConfig,
			Gosnomeru:  a.Config.GosnomeruConfig,
			Anomera:    a.Config.AnomeraConfig,
			Logger:     a.Logger,
		}); err != nil {
			return fmt.Errorf("register crons: %w", err)
		}

		group, groupCtx := errgroup.WithContext(ctx)

		group.Go(func() error {
			return sched.Run(groupCtx)
		})

		return group.Wait()
	})
}

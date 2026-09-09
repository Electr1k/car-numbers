package main

import (
	"context"
	"data-service/internal/app"
	"data-service/internal/feature"
	httptransport "data-service/internal/http"
	"data-service/internal/http/handler/feed"
	"data-service/internal/http/handler/region"
	"data-service/internal/job"
	"data-service/internal/job/cron"
	"data-service/internal/repository/postgres"
	"data-service/internal/scheduler"
	"data-service/internal/usecase/fetchfeednumbers"
	"data-service/internal/usecase/fetchregions"
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

		regionHandler := region.New(fetchregions.New(postgres.NewRegionRepository(a.Database)), a.Logger)
		feedHandler := feed.New(fetchfeednumbers.New(postgres.NewOfferRepository(a.Database)), a.Logger)
		router := httptransport.NewRouter(regionHandler, feedHandler)
		httpServer := httptransport.NewServer(a.Config.HttpServer, router)

		group, groupCtx := errgroup.WithContext(ctx)

		group.Go(func() error {
			return sched.Run(groupCtx)
		})

		group.Go(func() error {
			return httpServer.Run()
		})

		group.Go(func() error {
			<-groupCtx.Done()

			shutdownCtx, cancel := context.WithTimeout(context.Background(), a.Config.ShutdownTimeout)
			defer cancel()

			return httpServer.Shutdown(shutdownCtx)
		})

		return group.Wait()
	})
}

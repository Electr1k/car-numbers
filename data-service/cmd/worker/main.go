package main

import (
	"context"
	"data-service/internal/app"
	"data-service/internal/domain"
	"data-service/internal/feature"
	"data-service/internal/job"
	"data-service/internal/job/consumer"
	"data-service/internal/provider/resolver"
	"data-service/internal/repository/postgres"
	"data-service/internal/worker"
)

func main() {
	app.Run("worker jobs", func(ctx context.Context, a *app.App) error {
		jobRepository := postgres.NewJobRepository(a.Database)
		features := feature.NewFeature(postgres.NewFeatureRepository(a.Database), a.Logger)

		consumers := consumer.NewResolver()
		consumer.Register(consumers, consumer.Deps{
			Producer:   job.NewProducer(jobRepository, features),
			Providers:  resolver.New(a.Config, a.Logger),
			Offers:     postgres.NewOfferRepository(a.Database),
			Features:   features,
			AutoNomera: a.Config.AutoNomeraConfig,
			Logger:     a.Logger,
		})

		queues, err := domain.ParseJobQueues(a.Config.WorkerConfig.Queues)
		if err != nil {
			return err
		}

		return worker.New(
			jobRepository,
			consumers,
			queues,
			a.Config.WorkerConfig.Concurrency,
			a.Logger,
		).Run(ctx)
	})
}

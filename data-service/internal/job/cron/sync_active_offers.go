package cron

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/job"
	"fmt"
	"log/slog"
)

const nameSyncActiveOffers = "sync-active-offers"

// syncActiveOffers - ставит в очередь синхронизацию активных офферов
type syncActiveOffers struct {
	producer *job.Producer
	logger   *slog.Logger
}

func (c *syncActiveOffers) Run(ctx context.Context) error {
	for _, provider := range domain.GetAllProviders() {
		created, err := c.producer.DispatchSyncActiveOffers(ctx, provider)
		if err != nil {
			return fmt.Errorf("dispatch sync active offers for provider %s: %w", provider, err)
		}

		if !created {
			c.logger.Info("job skipped, previous one is still in the queue", "provider", provider)
		}
	}

	return nil
}

package cron

import (
	"context"
	"fmt"
	"log/slog"
	"plate-service/internal/domain"
	"plate-service/internal/job"
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

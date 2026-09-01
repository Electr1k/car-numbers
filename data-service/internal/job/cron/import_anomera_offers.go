package cron

import (
	"context"
	"data-service/internal/job"
	"fmt"
	"log/slog"
	"time"
)

const nameImportAnomeraOffers = "import-anomera-offers"

// importAnomeraOffers - ставит в очередь импорт офферов anomera
type importAnomeraOffers struct {
	producer *job.Producer
	depth    time.Duration
	logger   *slog.Logger
}

func (c *importAnomeraOffers) Run(ctx context.Context) error {
	created, err := c.producer.DispatchImportAnomeraOffers(ctx, job.ImportAnomeraOffersPayload{
		StopAfter: job.Duration(c.depth),
	})
	if err != nil {
		return fmt.Errorf("dispatch import anomera offers: %w", err)
	}

	if !created {
		c.logger.Info("job skipped, previous one is still in the queue")
	}

	return nil
}

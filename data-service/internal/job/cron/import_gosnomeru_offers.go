package cron

import (
	"context"
	"data-service/internal/job"
	"fmt"
	"log/slog"
	"time"
)

const nameImportGosnomeruOffers = "import-gosnomeru-offers"

// importGosnomeruOffers - ставит в очередь импорт офферов gosnomeru
type importGosnomeruOffers struct {
	producer *job.Producer
	depth    time.Duration
	logger   *slog.Logger
}

func (c *importGosnomeruOffers) Run(ctx context.Context) error {
	created, err := c.producer.DispatchImportGosnomeruOffers(ctx, job.ImportGosnomeruOffersPayload{
		StopAfter: job.Duration(c.depth),
	})
	if err != nil {
		return fmt.Errorf("dispatch import gosnomeru offers: %w", err)
	}

	if !created {
		c.logger.Info("job skipped, previous one is still in the queue")
	}

	return nil
}

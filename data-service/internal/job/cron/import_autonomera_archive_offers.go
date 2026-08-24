package cron

import (
	"context"
	"data-service/internal/job"
	"data-service/internal/provider/autonomera"
	"fmt"
	"log/slog"
	"time"
)

const nameImportAutonomeraArchiveOffers = "import-autonomera-archive-offers"

// importAutonomeraArchiveOffers - ставит в очередь импорт архивных офферов autonomera
type importAutonomeraArchiveOffers struct {
	producer *job.Producer
	depth    time.Duration
	logger   *slog.Logger
}

func (c *importAutonomeraArchiveOffers) Run(ctx context.Context) error {
	created, err := c.producer.DispatchImportAutonomeraOffers(ctx, job.ImportAutonomeraOffersPayload{
		Section:   autonomera.SectionArchive,
		StopAfter: job.Duration(c.depth),
	})
	if err != nil {
		return fmt.Errorf("dispatch import autonomera offers: %w", err)
	}

	if !created {
		c.logger.Info("job skipped, previous one is still in the queue")
	}

	return nil
}

package cron

import (
	"context"
	"fmt"
	"log/slog"
	"plate-service/internal/job"
	"plate-service/internal/provider/autonomera"
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

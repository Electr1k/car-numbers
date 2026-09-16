package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"plate-service/internal/job"
	"plate-service/internal/usecase/importautonomera"
	"time"
)

type ImportAutonomeraOffersConsumer struct {
	uc *importautonomera.UseCase
}

func NewImportAutonomeraOffersConsumer(uc *importautonomera.UseCase) *ImportAutonomeraOffersConsumer {
	return &ImportAutonomeraOffersConsumer{
		uc: uc,
	}
}

func (c *ImportAutonomeraOffersConsumer) Handle(ctx context.Context, payload string) error {
	var decoded job.ImportAutonomeraOffersPayload

	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	return c.uc.Handle(ctx, importautonomera.Params{
		Section:     decoded.Section,
		StartOffset: decoded.StartOffset,
		MaxPages:    decoded.MaxPages,
		StopAfter:   time.Duration(decoded.StopAfter),
	})
}

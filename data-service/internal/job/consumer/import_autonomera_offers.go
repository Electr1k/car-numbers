package consumer

import (
	"context"
	"data-service/internal/job"
	"data-service/internal/usecase"
	"encoding/json"
	"fmt"
	"time"
)

type ImportAutonomeraOffersConsumer struct {
	uc *usecase.ImportAutonomeraOffersUseCase
}

func NewImportAutonomeraOffersConsumer(uc *usecase.ImportAutonomeraOffersUseCase) *ImportAutonomeraOffersConsumer {
	return &ImportAutonomeraOffersConsumer{
		uc: uc,
	}
}

func (c *ImportAutonomeraOffersConsumer) Handle(ctx context.Context, payload string) error {
	var decoded job.ImportAutonomeraOffersPayload

	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	return c.uc.Handle(ctx, usecase.ImportParams{
		Section:     decoded.Section,
		StartOffset: decoded.StartOffset,
		MaxPages:    decoded.MaxPages,
		StopAfter:   time.Duration(decoded.StopAfter),
	})
}

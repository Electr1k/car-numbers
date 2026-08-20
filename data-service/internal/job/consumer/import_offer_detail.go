package consumer

import (
	"context"
	"data-service/internal/job"
	"data-service/internal/usecase"
	"encoding/json"
	"fmt"
)

type ImportOfferDetailConsumer struct {
	uc *usecase.ImportOfferDetailUseCase
}

func NewImportOfferDetailConsumer(uc *usecase.ImportOfferDetailUseCase) *ImportOfferDetailConsumer {
	return &ImportOfferDetailConsumer{
		uc: uc,
	}
}

func (c *ImportOfferDetailConsumer) Handle(ctx context.Context, payload string) error {
	var decoded job.ImportOfferDetailPayload

	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	return c.uc.Handle(ctx, decoded.OfferId)
}

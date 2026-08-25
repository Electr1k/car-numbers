package consumer

import (
	"context"
	"data-service/internal/job"
	"data-service/internal/usecase/importofferdetail"
	"encoding/json"
	"fmt"
)

type ImportOfferDetailConsumer struct {
	uc *importofferdetail.UseCase
}

func NewImportOfferDetailConsumer(uc *importofferdetail.UseCase) *ImportOfferDetailConsumer {
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

package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"plate-service/internal/job"
	"plate-service/internal/usecase/importofferdetail"
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

	return c.uc.Handle(ctx, decoded.OfferID)
}

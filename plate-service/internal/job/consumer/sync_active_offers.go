package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"plate-service/internal/job"
	"plate-service/internal/usecase/syncactiveoffers"
)

type SyncActiveOffersConsumer struct {
	uc *syncactiveoffers.UseCase
}

func NewSyncActiveOffersConsumer(uc *syncactiveoffers.UseCase) *SyncActiveOffersConsumer {
	return &SyncActiveOffersConsumer{
		uc: uc,
	}
}

func (c *SyncActiveOffersConsumer) Handle(ctx context.Context, payload string) error {
	var decoded job.SyncActiveOffersPayload

	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	return c.uc.Handle(ctx, syncactiveoffers.Params{
		Provider: decoded.Provider,
	})
}

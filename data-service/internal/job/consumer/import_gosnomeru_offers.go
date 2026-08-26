package consumer

import (
	"context"
	"data-service/internal/job"
	"data-service/internal/usecase/importgosnomeru"
	"encoding/json"
	"fmt"
	"time"
)

type ImportGosnomeruOffersConsumer struct {
	uc *importgosnomeru.UseCase
}

func NewImportGosnomeruOffersConsumer(uc *importgosnomeru.UseCase) *ImportGosnomeruOffersConsumer {
	return &ImportGosnomeruOffersConsumer{
		uc: uc,
	}
}

func (c *ImportGosnomeruOffersConsumer) Handle(ctx context.Context, payload string) error {
	var decoded job.ImportGosnomeruOffersPayload

	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	return c.uc.Handle(ctx, importgosnomeru.Params{
		StartPage: decoded.StartPage,
		MaxPages:  decoded.MaxPages,
		StopAfter: time.Duration(decoded.StopAfter),
	})
}

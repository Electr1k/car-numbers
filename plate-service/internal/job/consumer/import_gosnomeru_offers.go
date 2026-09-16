package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"plate-service/internal/job"
	"plate-service/internal/usecase/importgosnomeru"
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

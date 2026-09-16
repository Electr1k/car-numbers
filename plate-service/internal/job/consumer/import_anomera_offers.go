package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"plate-service/internal/job"
	"plate-service/internal/usecase/importanomera"
	"time"
)

type ImportAnomeraOffersConsumer struct {
	uc *importanomera.UseCase
}

func NewImportAnomeraOffersConsumer(uc *importanomera.UseCase) *ImportAnomeraOffersConsumer {
	return &ImportAnomeraOffersConsumer{
		uc: uc,
	}
}

func (c *ImportAnomeraOffersConsumer) Handle(ctx context.Context, payload string) error {
	var decoded job.ImportAnomeraOffersPayload

	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	return c.uc.Handle(ctx, importanomera.Params{
		StartPage: decoded.StartPage,
		MaxPages:  decoded.MaxPages,
		StopAfter: time.Duration(decoded.StopAfter),
	})
}

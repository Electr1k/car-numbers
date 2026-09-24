package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"plate-service/internal/job"
	"plate-service/internal/usecase/importprofile"
)

type ImportProfileConsumer struct {
	uc *importprofile.UseCase
}

func NewImportProfileConsumer(uc *importprofile.UseCase) *ImportProfileConsumer {
	return &ImportProfileConsumer{
		uc: uc,
	}
}

func (c *ImportProfileConsumer) Handle(ctx context.Context, payload string) error {
	var decoded job.ImportProfilePayload

	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	return c.uc.Handle(ctx, importprofile.Params{Provider: decoded.Provider, ProfileExternalID: decoded.ProfileExternalId})
}

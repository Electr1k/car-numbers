package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"plate-service/internal/job"
	"plate-service/internal/usecase/importautonomeraprofiles"
)

type ImportAutonomeraProfilesConsumer struct {
	uc *importautonomeraprofiles.UseCase
}

func NewImportAutonomeraProfilesConsumer(uc *importautonomeraprofiles.UseCase) *ImportAutonomeraProfilesConsumer {
	return &ImportAutonomeraProfilesConsumer{
		uc: uc,
	}
}

func (c *ImportAutonomeraProfilesConsumer) Handle(ctx context.Context, payload string) error {
	var decoded job.ImportAutonomeraProfilesPayload

	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	return c.uc.Handle(ctx)
}

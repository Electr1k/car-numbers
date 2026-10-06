package job

import (
	"context"
	"encoding/json"
	"fmt"
	"plate-service/internal/domain"
)

type ImportAutonomeraProfilesPayload struct {
}

func importAutonomeraProfilesUniqueKey() string {
	return fmt.Sprintf("%s", domain.JobNameImportAutonomeraProfiles)
}

func (p *Producer) DispatchImportAutonomeraProfiles(ctx context.Context, payload ImportAutonomeraProfilesPayload) (bool, error) {
	enabled, err := p.features.Enabled(ctx, domain.FeatureKeyDispatchImportAutonomeraProfiles)
	if err != nil {
		return false, err
	}
	if !enabled {
		return false, nil
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("marshal payload: %w", err)
	}

	return p.dispatch(ctx, newTask(
		domain.JobNameImportAutonomeraProfiles,
		domain.JobQueueAutonomera,
		string(encoded),
		importAutonomeraProfilesUniqueKey(),
	))
}

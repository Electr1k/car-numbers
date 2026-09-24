package job

import (
	"context"
	"encoding/json"
	"fmt"
	"plate-service/internal/domain"
)

type ImportProfilePayload struct {
	Provider          domain.Provider `json:"provider"`
	ProfileExternalId string          `json:"profile_external_id"`
}

func importProfileUniqueKey(provider domain.Provider, profileExternalId string) string {
	return fmt.Sprintf("%s:%s-%s", domain.JobNameImportProfile, provider, profileExternalId)
}

func (p *Producer) DispatchImportProfile(ctx context.Context, provider domain.Provider, profileExternalId string) (bool, error) {
	enabled, err := p.features.Enabled(ctx, domain.FeatureKeyDispatchImportProfile)
	if err != nil {
		return false, err
	}
	if !enabled {
		return false, nil
	}

	encoded, err := json.Marshal(ImportProfilePayload{Provider: provider, ProfileExternalId: profileExternalId})
	if err != nil {
		return false, fmt.Errorf("marshal payload: %w", err)
	}

	return p.dispatch(ctx, newTask(
		domain.JobNameImportProfile,
		provider.JobQueue(),
		string(encoded),
		importProfileUniqueKey(provider, profileExternalId),
	))
}

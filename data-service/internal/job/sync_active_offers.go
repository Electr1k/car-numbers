package job

import (
	"context"
	"data-service/internal/domain"
	"encoding/json"
	"fmt"
)

type SyncActiveOffersPayload struct {
	Provider domain.Provider `json:"provider"`
}

func syncActiveOffersUniqueKey(provider domain.Provider) string {
	return fmt.Sprintf("%s:%s", domain.JobNameSyncActiveOffers, provider)
}

func (p *Producer) DispatchSyncActiveOffers(ctx context.Context, provider domain.Provider) (bool, error) {
	enabled, err := p.features.Enabled(ctx, domain.FeatureKeyDispatchSyncActiveOffers)
	if err != nil {
		return false, err
	}
	if !enabled {
		return false, nil
	}

	encoded, err := json.Marshal(SyncActiveOffersPayload{Provider: provider})
	if err != nil {
		return false, fmt.Errorf("marshal payload: %w", err)
	}

	return p.dispatch(ctx, newTask(
		domain.JobNameSyncActiveOffers,
		provider.JobQueue(),
		string(encoded),
		syncActiveOffersUniqueKey(provider),
	))
}

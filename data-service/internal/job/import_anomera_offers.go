package job

import (
	"context"
	"data-service/internal/domain"
	"encoding/json"
	"fmt"
)

type ImportAnomeraOffersPayload struct {
	StartPage int      `json:"start_page"`
	MaxPages  int      `json:"max_pages"`
	StopAfter Duration `json:"stop_after"`
}

func importAnomeraOffersUniqueKey() string {
	return string(domain.JobNameImportAnomeraOffers)
}

func (p *Producer) DispatchImportAnomeraOffers(ctx context.Context, payload ImportAnomeraOffersPayload) (bool, error) {
	enabled, err := p.features.Enabled(ctx, domain.FeatureKeyDispatchImportAnomeraOffers)
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
		domain.JobNameImportAnomeraOffers,
		domain.JobQueueAnomera,
		string(encoded),
		importAnomeraOffersUniqueKey(),
	))
}

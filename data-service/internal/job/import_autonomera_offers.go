package job

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/provider/autonomera"
	"encoding/json"
	"fmt"
)

type ImportAutonomeraOffersPayload struct {
	Section     autonomera.Section `json:"section"`
	StartOffset int                `json:"start_offset"`
	MaxPages    int                `json:"max_pages"`
	StopAfter   Duration           `json:"stop_after"`
}

func importAutonomeraOffersUniqueKey(section autonomera.Section) string {
	return fmt.Sprintf("%s:%s", domain.JobNameImportAutonomeraOffers, section)
}

func (p *Producer) DispatchImportAutonomeraOffers(ctx context.Context, payload ImportAutonomeraOffersPayload) (bool, error) {
	enabled, err := p.features.Enabled(ctx, domain.FeatureKeyDispatchImportAutonomeraOffers)
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
		domain.JobNameImportAutonomeraOffers,
		domain.JobQueueAutonomera,
		string(encoded),
		importAutonomeraOffersUniqueKey(payload.Section),
	))
}

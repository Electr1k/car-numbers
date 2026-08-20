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
	encoded, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("marshal payload: %w", err)
	}

	task, err := newTask(
		domain.JobNameImportAutonomeraOffers,
		domain.JobQueueImportAutonomeraOffers,
		string(encoded),
		importAutonomeraOffersUniqueKey(payload.Section),
	)
	if err != nil {
		return false, fmt.Errorf("new task: %w", err)
	}

	return p.dispatch(ctx, task)
}

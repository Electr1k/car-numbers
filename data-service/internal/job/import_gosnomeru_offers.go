package job

import (
	"context"
	"data-service/internal/domain"
	"encoding/json"
	"fmt"
)

type ImportGosnomeruOffersPayload struct {
	StartPage int      `json:"start_page"`
	MaxPages  int      `json:"max_pages"`
	StopAfter Duration `json:"stop_after"`
}

func importGosnomeruOffersUniqueKey() string {
	return string(domain.JobNameImportGosnomeruOffers)
}

func (p *Producer) DispatchImportGosnomeruOffers(ctx context.Context, payload ImportGosnomeruOffersPayload) (bool, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("marshal payload: %w", err)
	}

	return p.dispatch(ctx, newTask(
		domain.JobNameImportGosnomeruOffers,
		domain.JobQueueImportGosnomeruOffers,
		string(encoded),
		importGosnomeruOffersUniqueKey(),
	))
}

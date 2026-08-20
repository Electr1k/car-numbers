package job

import (
	"context"
	"data-service/internal/domain"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type ImportOfferDetailPayload struct {
	OfferId uuid.UUID `json:"offer_id"`
}

func importOfferDetailUniqueKey(offerId uuid.UUID) string {
	return fmt.Sprintf("%s:%s", domain.JobNameImportOfferDetail, offerId)
}

func (p *Producer) DispatchImportOfferDetail(ctx context.Context, payload ImportOfferDetailPayload) (bool, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("marshal payload: %w", err)
	}

	task, err := newTask(
		domain.JobNameImportOfferDetail,
		domain.JobQueueImportOfferDetail,
		string(encoded),
		importOfferDetailUniqueKey(payload.OfferId),
	)
	if err != nil {
		return false, fmt.Errorf("new task: %w", err)
	}

	return p.dispatch(ctx, task)
}

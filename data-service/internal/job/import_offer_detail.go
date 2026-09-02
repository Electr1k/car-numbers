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

func (p *Producer) DispatchImportOfferDetail(ctx context.Context, offerId uuid.UUID, provider domain.Provider) (bool, error) {
	enabled, err := p.features.Enabled(ctx, domain.FeatureKeyDispatchImportOfferDetail)
	if err != nil {
		return false, err
	}
	if !enabled {
		return false, nil
	}

	encoded, err := json.Marshal(ImportOfferDetailPayload{OfferId: offerId})
	if err != nil {
		return false, fmt.Errorf("marshal payload: %w", err)
	}

	return p.dispatch(ctx, newTask(
		domain.JobNameImportOfferDetail,
		provider.JobQueue(),
		string(encoded),
		importOfferDetailUniqueKey(offerId),
	))
}

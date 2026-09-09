package job

import (
	"context"
	"data-service/internal/domain"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

type ImportOfferDetailPayload struct {
	OfferID uuid.UUID `json:"offer_id"`
}

func importOfferDetailUniqueKey(offerID uuid.UUID) string {
	return fmt.Sprintf("%s:%s", domain.JobNameImportOfferDetail, offerID)
}

func (p *Producer) DispatchImportOfferDetail(ctx context.Context, offerID uuid.UUID, provider domain.Provider) (bool, error) {
	enabled, err := p.features.Enabled(ctx, domain.FeatureKeyDispatchImportOfferDetail)
	if err != nil {
		return false, err
	}
	if !enabled {
		return false, nil
	}

	encoded, err := json.Marshal(ImportOfferDetailPayload{OfferID: offerID})
	if err != nil {
		return false, fmt.Errorf("marshal payload: %w", err)
	}

	return p.dispatch(ctx, newTask(
		domain.JobNameImportOfferDetail,
		provider.JobQueue(),
		string(encoded),
		importOfferDetailUniqueKey(offerID),
	))
}

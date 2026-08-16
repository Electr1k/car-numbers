package job

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/usecase"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ImportOfferDetailJob struct {
	job *Job
	uc  *usecase.ImportOfferDetailUseCase
}

type ImportOfferDetailJobPayload struct {
	OfferId uuid.UUID `json:"offer_id"`
}

func NewImportOfferDetailJob(store store, uc *usecase.ImportOfferDetailUseCase) *ImportOfferDetailJob {
	return &ImportOfferDetailJob{
		job: newJob(store, domain.JobNameImportOfferDetailJobName, domain.JobQueueDefault, false),
		uc:  uc,
	}
}

func ImportOfferDetailUniqueKey(offerId uuid.UUID) string {
	return fmt.Sprintf("%s:%s", domain.JobNameImportOfferDetailJobName, offerId)
}

func (j *ImportOfferDetailJob) Dispatch(ctx context.Context, payload ImportOfferDetailJobPayload, startAfter *time.Time) (bool, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("marshal payload: %w", err)
	}

	created, err := j.job.dispatch(ctx, string(encoded), ImportOfferDetailUniqueKey(payload.OfferId), startAfter)
	if err != nil {
		return false, fmt.Errorf("dispatch job: %w", err)
	}

	return created, nil
}

func (j *ImportOfferDetailJob) Handle(ctx context.Context, payload string) error {
	var decoded ImportOfferDetailJobPayload

	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	return j.uc.Handle(ctx, decoded.OfferId)
}

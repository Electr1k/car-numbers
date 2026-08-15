package job

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/provider/autonomera"
	"data-service/internal/usecase"
	"encoding/json"
	"fmt"
	"time"
)

type ImportAutonomeraJob struct {
	job *Job
	uc  *usecase.ImportAutonomeraOffersUseCase
}

type ImportAutonomeraPayload struct {
	Section     autonomera.Section `json:"section"`
	StartOffset int                `json:"start_offset"`
	MaxPages    int                `json:"max_pages"`
	StopAfter   Duration           `json:"stop_after"`
}

func NewImportAutonomeraJob(store store, uc *usecase.ImportAutonomeraOffersUseCase) *ImportAutonomeraJob {
	return &ImportAutonomeraJob{
		job: newJob(store, domain.JobNameImportAutonomeraOffers, domain.JobQueueDefault, false),
		uc:  uc,
	}
}

func ImportAutonomeraUniqueKey(section autonomera.Section) string {
	return fmt.Sprintf("%s:%s", domain.JobNameImportAutonomeraOffers, section)
}

func (j *ImportAutonomeraJob) Dispatch(ctx context.Context, payload ImportAutonomeraPayload, startAfter *time.Time) (bool, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("marshal payload: %w", err)
	}

	created, err := j.job.dispatch(ctx, string(encoded), ImportAutonomeraUniqueKey(payload.Section), startAfter)
	if err != nil {
		return false, fmt.Errorf("dispatch job: %w", err)
	}

	return created, nil
}

func (j *ImportAutonomeraJob) Handle(ctx context.Context, payload string) error {
	var decoded ImportAutonomeraPayload

	if err := json.Unmarshal([]byte(payload), &decoded); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	return j.uc.Handle(ctx, usecase.ImportParams{
		Section:     decoded.Section,
		StartOffset: decoded.StartOffset,
		MaxPages:    decoded.MaxPages,
		StopAfter:   time.Duration(decoded.StopAfter),
	})
}

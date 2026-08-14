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
	uc  usecase.ImportAutonomeraOffersUseCase
}

type ImportAutonomeraPayload struct {
	Section     autonomera.Section `json:"section"`
	StartOffset int                `json:"startOffset"`
	MaxPages    int                `json:"maxPages"`
	StopAfter   time.Duration      `json:"stopAfter"`
}

func NewImportAutonomeraJob(repository storeInterface, usecase usecase.ImportAutonomeraOffersUseCase) (*ImportAutonomeraJob, error) {
	return &ImportAutonomeraJob{
		job: newJob(repository, domain.JobNameImportAutonomerOffers, domain.JobQueueDefault, false),
		uc:  usecase,
	}, nil
}

func (j *ImportAutonomeraJob) Dispatch(ctx context.Context, payload ImportAutonomeraPayload, startAfter *time.Time) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	err = j.job.dispatch(ctx, string(jsonData), startAfter)
	if err != nil {
		return fmt.Errorf("failed to dispatch job: %w", err)
	}

	return nil
}

func (i ImportAutonomeraJob) Handle(ctx context.Context, payload string) error {
	var payloadObject ImportAutonomeraPayload

	err := json.Unmarshal([]byte(payload), &payloadObject)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	err = i.uc.Handle(ctx, usecase.ImportParams{
		Section:     payloadObject.Section,
		StartOffset: payloadObject.StartOffset,
		MaxPages:    payloadObject.MaxPages,
		StopAfter:   payloadObject.StopAfter,
	})
	if err != nil {
		return err
	}

	return nil
}

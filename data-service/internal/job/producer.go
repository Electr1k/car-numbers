package job

import (
	"context"
	"data-service/internal/domain"
	"fmt"
	"time"
)

type store interface {
	CreateJob(ctx context.Context, job domain.Job) (bool, error)
}

type feature interface {
	Enabled(ctx context.Context, key domain.FeatureKey) (bool, error)
}

type Producer struct {
	jobStore store
	features feature
}

func NewProducer(jobStore store, features feature) *Producer {
	return &Producer{
		jobStore: jobStore,
		features: features,
	}
}

func (p *Producer) dispatch(ctx context.Context, t *task) (bool, error) {
	domainJob, err := domain.NewJob(
		t.name,
		t.queue,
		domain.JobStatusPending,
		time.Now(),
		t.payload,
		t.uniqueKey,
	)
	if err != nil {
		return false, fmt.Errorf("create domain job: %w", err)
	}

	created, err := p.jobStore.CreateJob(ctx, *domainJob)
	if err != nil {
		return false, fmt.Errorf("dispatch job: %w", err)
	}

	return created, nil
}

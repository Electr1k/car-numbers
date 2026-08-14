package job

import (
	"context"
	"data-service/internal/domain"
	"fmt"
)

type RepositoryInterface interface {
	GetJob(ctx context.Context) (*domain.Job, error)
}

type JobInterface interface {
	Handle(ctx context.Context, payload string) error
}

type Resolver struct {
	repository          RepositoryInterface
	importAutonomeraJob *ImportAutonomeraJob
}

func NewResolver(
	repository RepositoryInterface,
	importAutonomeraJob *ImportAutonomeraJob,
) *Resolver {
	return &Resolver{repository: repository, importAutonomeraJob: importAutonomeraJob}
}

func (m *Resolver) ResolveJob(job domain.Job) (JobInterface, error) {
	switch job.Name {
	case domain.JobNameImportAutonomerOffers:
		return m.importAutonomeraJob, nil
	default:
		return nil, fmt.Errorf("unknown job name: %s", job.Name)
	}
}

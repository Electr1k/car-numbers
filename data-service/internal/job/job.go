package job

import (
	"context"
	"data-service/internal/domain"
	"fmt"
	"time"
)

type store interface {
	CreateJob(ctx context.Context, job domain.Job) (*domain.Job, error)
}

type Job struct {
	store store

	name domain.JobName

	queue domain.JobQueue

	executedParallel bool
}

func newJob(
	store store,
	name domain.JobName,
	queue domain.JobQueue,
	executedParallel bool,
) *Job {
	return &Job{
		store:            store,
		name:             name,
		queue:            queue,
		executedParallel: executedParallel,
	}
}

func (j *Job) dispatch(ctx context.Context, payload string, startAfter *time.Time) error {

	if startAfter == nil {
		start := time.Now()
		startAfter = &start
	}

	domainJob, err := domain.NewJob(j.name, j.queue, domain.JobStatusPending, *startAfter, payload)
	if err != nil {
		return fmt.Errorf("dispatch job: %w", err)
	}

	_, err = j.store.CreateJob(ctx, *domainJob)
	if err != nil {
		return fmt.Errorf("dispatch job: %w", err)
	}

	return nil
}

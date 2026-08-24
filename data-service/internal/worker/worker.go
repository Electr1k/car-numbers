package worker

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/job/consumer"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

const (
	pollInterval = time.Second

	finishTimeout = 5 * time.Second
)

type JobStore interface {
	GetJob(ctx context.Context, queues []domain.JobQueue) (*domain.Job, error)
	MarkJobFailed(ctx context.Context, id uuid.UUID, jobErr error) error
	DeleteJob(ctx context.Context, id uuid.UUID) error
}

type ConsumerResolver interface {
	Resolve(name domain.JobName) (consumer.Handler, error)
}

type Worker struct {
	store    JobStore
	resolver ConsumerResolver
	queues   []domain.JobQueue
	logger   *slog.Logger
}

func New(store JobStore, resolver ConsumerResolver, queues []domain.JobQueue, logger *slog.Logger) *Worker {
	return &Worker{
		store:    store,
		resolver: resolver,
		queues:   queues,
		logger:   logger.With("queues", queues),
	}
}

func (w *Worker) Run(ctx context.Context) error {
	w.logger.Info("worker started")

	for {
		if ctx.Err() != nil {
			w.logger.Info("worker stopped")
			return nil
		}

		if err := w.processNext(ctx); err != nil {
			w.logger.Error("process job", "error", err)
			time.Sleep(pollInterval)
		}
	}
}

func (w *Worker) processNext(ctx context.Context) error {
	domainJob, err := w.store.GetJob(ctx, w.queues)
	if errors.Is(err, domain.ErrNoJob) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("get job: %w", err)
	}

	logger := w.logger.With("job_id", domainJob.Id, "job", domainJob.Name, "queue", domainJob.Queue)
	logger.Info("job taken")

	started := time.Now()
	handleErr := w.handle(ctx, *domainJob)

	if handleErr != nil && ctx.Err() != nil {
		logger.Info("job interrupted by shutdown, will be retaken after lock expires")
		return nil
	}

	finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), finishTimeout)
	defer cancel()

	if handleErr != nil {
		logger.Error("job failed", "error", handleErr, "duration", time.Since(started))

		if err := w.store.MarkJobFailed(finishCtx, domainJob.Id, handleErr); err != nil {
			return fmt.Errorf("mark job failed: %w", err)
		}

		return nil
	}

	logger.Info("job finished", "duration", time.Since(started))

	if err := w.store.DeleteJob(finishCtx, domainJob.Id); err != nil {
		return fmt.Errorf("delete job: %w", err)
	}

	return nil
}

func (w *Worker) handle(ctx context.Context, domainJob domain.Job) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("job panicked: %v\n%s", r, debug.Stack())
		}
	}()

	handler, err := w.resolver.Resolve(domainJob.Name)
	if err != nil {
		return err
	}

	return handler.Handle(ctx, domainJob.Payload)
}

package worker

import (
	"context"
	"data-service/internal/domain"
	"data-service/internal/job/consumer"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
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
	store       JobStore
	resolver    ConsumerResolver
	queues      []domain.JobQueue
	concurrency int
	logger      *slog.Logger
}

func New(
	store JobStore,
	resolver ConsumerResolver,
	queues []domain.JobQueue,
	concurrency int,
	logger *slog.Logger,
) *Worker {
	if concurrency < 1 {
		concurrency = 1
	}

	return &Worker{
		store:       store,
		resolver:    resolver,
		queues:      queues,
		concurrency: concurrency,
		logger:      logger.With("queues", queues),
	}
}

// Run - запускает в горутинах воркеры, которые слушают очереди
func (w *Worker) Run(ctx context.Context) error {
	w.logger.Info("worker started", "concurrency", w.concurrency)

	group, groupCtx := errgroup.WithContext(ctx)
	for i := range w.concurrency {
		group.Go(func() error {
			w.loop(groupCtx, w.logger.With("consumer", i))

			return nil
		})
	}

	err := group.Wait()
	w.logger.Info("worker stopped")

	return err
}

// loop - разбор очередей до отмены контекста
func (w *Worker) loop(ctx context.Context, logger *slog.Logger) {
	for {
		if ctx.Err() != nil {
			return
		}

		taken, err := w.processNext(ctx, logger)
		switch {
		case err != nil:
			logger.Error("process job", "error", err)
			w.idle(ctx)
		case !taken:
			w.idle(ctx)
		}
	}
}

// idle - пауза перед следующим опросом
func (w *Worker) idle(ctx context.Context) {
	timer := time.NewTimer(pollInterval/2 + rand.N(pollInterval))
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// processNext - процессит джобу
func (w *Worker) processNext(ctx context.Context, workerLogger *slog.Logger) (bool, error) {
	domainJob, err := w.store.GetJob(ctx, w.queues)
	if errors.Is(err, domain.ErrNoJob) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get job: %w", err)
	}

	logger := workerLogger.With("job_id", domainJob.Id, "job", domainJob.Name, "queue", domainJob.Queue)
	logger.Info("job taken")

	started := time.Now()
	handleErr := w.handle(ctx, *domainJob)

	if handleErr != nil && ctx.Err() != nil {
		logger.Info("job interrupted by shutdown, will be retaken after lock expires")
		return true, nil
	}

	finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), finishTimeout)
	defer cancel()

	if handleErr != nil {
		logger.Error("job failed", "error", handleErr, "duration", time.Since(started))

		if err := w.store.MarkJobFailed(finishCtx, domainJob.Id, handleErr); err != nil {
			return true, fmt.Errorf("mark job failed: %w", err)
		}

		return true, nil
	}

	logger.Info("job finished", "duration", time.Since(started))

	if err := w.store.DeleteJob(finishCtx, domainJob.Id); err != nil {
		return true, fmt.Errorf("delete job: %w", err)
	}

	return true, nil
}

// handle - запускает джобу и ловит пиники
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

package scheduler

import (
	"context"
	"data-service/internal/domain"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

const tickOffset = 100 * time.Millisecond

type JobStore interface {
	CreateJob(ctx context.Context, job domain.Job) (bool, error)
}

type Entry struct {
	Name      domain.JobName
	Queue     domain.JobQueue
	Spec      string
	Payload   string
	UniqueKey string
}

type entry struct {
	Entry
	schedule cron.Schedule
}

func (e entry) due(now time.Time) bool {
	return e.schedule.Next(now.Add(-time.Second)).Equal(now)
}

type Scheduler struct {
	store   JobStore
	logger  *slog.Logger
	entries []entry
}

func New(store JobStore, logger *slog.Logger) *Scheduler {
	return &Scheduler{store: store, logger: logger}
}

func (s *Scheduler) Add(e Entry) error {
	switch {
	case e.Name == "":
		return errors.New("entry name must not be empty")
	case e.Queue == "":
		return fmt.Errorf("entry %q: queue must not be empty", e.Name)
	case e.Payload == "":
		return fmt.Errorf("entry %q: payload must not be empty", e.Name)
	case e.UniqueKey == "":
		return fmt.Errorf("entry %q: unique key must not be empty", e.Name)
	}

	for _, existing := range s.entries {
		if existing.UniqueKey == e.UniqueKey {
			return fmt.Errorf("entry with unique key %q is already registered", e.UniqueKey)
		}
	}

	schedule, err := cron.ParseStandard(e.Spec)
	if err != nil {
		return fmt.Errorf("entry %q: parse spec %q: %w", e.Name, e.Spec, err)
	}

	s.entries = append(s.entries, entry{Entry: e, schedule: schedule})

	return nil
}

func (s *Scheduler) Run(ctx context.Context) error {
	if len(s.entries) == 0 {
		return errors.New("no entries registered")
	}

	s.logger.Info("scheduler started", "entries", len(s.entries))

	for {
		if !sleep(ctx, time.Until(nextTick(time.Now()))) {
			s.logger.Info("scheduler stopped")
			return nil
		}

		s.tick(ctx, time.Now().Truncate(time.Minute))
	}
}

func (s *Scheduler) tick(ctx context.Context, now time.Time) {
	for _, e := range s.entries {
		if !e.due(now) {
			continue
		}

		logger := s.logger.With("job", e.Name, "queue", e.Queue, "spec", e.Spec)

		domainJob, err := domain.NewJob(e.Name, e.Queue, domain.JobStatusPending, now, e.Payload, e.UniqueKey)
		if err != nil {
			logger.Error("build scheduled job", "error", err)
			continue
		}

		created, err := s.store.CreateJob(ctx, *domainJob)
		if err != nil {
			logger.Error("enqueue scheduled job", "error", err)
			continue
		}

		if !created {
			logger.Info("job skipped, previous one is still in the queue")
			continue
		}

		logger.Info("job enqueued", "job_id", domainJob.Id)
	}
}

func nextTick(now time.Time) time.Time {
	return now.Truncate(time.Minute).Add(time.Minute + tickOffset)
}

func sleep(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}

	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

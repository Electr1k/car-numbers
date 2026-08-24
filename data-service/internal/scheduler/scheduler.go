package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

const runTimeout = time.Minute

type Task interface {
	Run(ctx context.Context) error
}

// entry - зарегистрированная задача со своим расписанием
type entry struct {
	name     string
	spec     string
	schedule cron.Schedule
	task     Task
}

// Scheduler - запускает задачи по cron-расписанию
type Scheduler struct {
	entries []entry
	logger  *slog.Logger
}

func New(logger *slog.Logger) *Scheduler {
	return &Scheduler{logger: logger}
}

// Add - регистрирует задачу, спека разбирается сразу, на старте процесса
func (s *Scheduler) Add(name string, spec string, task Task) error {
	switch {
	case name == "":
		return errors.New("cron name must not be empty")
	case task == nil:
		return fmt.Errorf("cron %q: task must not be nil", name)
	}

	schedule, err := cron.ParseStandard(spec)
	if err != nil {
		return fmt.Errorf("cron %q: parse spec %q: %w", name, spec, err)
	}

	s.entries = append(s.entries, entry{
		name:     name,
		spec:     spec,
		schedule: schedule,
		task:     task,
	})

	return nil
}

func (s *Scheduler) Run(ctx context.Context) error {
	if len(s.entries) == 0 {
		s.logger.Warn("scheduler has no entries, idling")
		<-ctx.Done()

		return nil
	}

	s.logger.Info("scheduler started", "entries", len(s.entries))

	var wg sync.WaitGroup

	for _, e := range s.entries {
		wg.Add(1)

		go func() {
			defer wg.Done()
			s.loop(ctx, e)
		}()
	}

	wg.Wait()
	s.logger.Info("scheduler stopped")

	return nil
}

// loop - крутит одну задачу: ждёт ближайший тик и запускает
func (s *Scheduler) loop(ctx context.Context, e entry) {
	logger := s.logger.With("cron", e.name, "spec", e.spec)

	for {
		next := e.schedule.Next(time.Now())
		logger.Debug("next run scheduled", "at", next)

		if !sleep(ctx, time.Until(next)) {
			return
		}

		started := time.Now()

		err := s.run(ctx, e)

		switch {
		case err != nil && ctx.Err() != nil:
			logger.Info("cron interrupted by shutdown")
			return
		case err != nil:
			logger.Error("cron failed", "error", err, "duration", time.Since(started))
		default:
			logger.Info("cron finished", "duration", time.Since(started))
		}

		if time.Now().After(e.schedule.Next(next)) {
			logger.Warn("cron run overran its next tick", "duration", time.Since(started))
		}
	}
}

// run - выполняет задачу под таймаутом, паника задачи не роняет шедулер
func (s *Scheduler) run(ctx context.Context, e entry) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cron panicked: %v\n%s", r, debug.Stack())
		}
	}()

	runCtx, cancel := context.WithTimeout(ctx, runTimeout)
	defer cancel()

	return e.task.Run(runCtx)
}

// sleep - ждёт d, false если контекст отменён
func sleep(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

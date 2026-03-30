package carnumber

import (
	"context"
	"core/internal/domain"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type NumberProvider interface {
	FetchNumbers(ctx context.Context) (domain.NumberIterator, error)
}

type CarNumberRepository interface {
	UpdateOrCreate(ctx context.Context, carNumber *domain.CarNumber) (*domain.CarNumber, error)
}

type FetchCarNumberUseCase struct {
	provider       NumberProvider
	repository     CarNumberRepository
	logger         *slog.Logger
	stopAfter      time.Duration
	rateLimitDelay time.Duration
}

type Config struct {
	StopAfter      time.Duration
	RateLimitDelay time.Duration
}

func NewFetchCarNumberUseCase(
	provider NumberProvider,
	repository CarNumberRepository,
	logger *slog.Logger,
	cfg Config,
) *FetchCarNumberUseCase {
	return &FetchCarNumberUseCase{
		provider:       provider,
		repository:     repository,
		logger:         logger,
		stopAfter:      cfg.StopAfter,
		rateLimitDelay: cfg.RateLimitDelay,
	}
}

func (uc *FetchCarNumberUseCase) Handle(ctx context.Context) error {
	uc.logger.Info("starting car number fetch",
		"stop_after", uc.stopAfter,
		"rate_limit_delay", uc.rateLimitDelay)

	iterator, err := uc.provider.FetchNumbers(ctx)
	if err != nil {
		uc.logger.Error("failed to create iterator", "error", err)
		return fmt.Errorf("create iterator: %w", err)
	}

	cutoffDate := time.Now().Add(-uc.stopAfter)
	uc.logger.Info("cutoff date calculated", "cutoff_date", cutoffDate)

	metrics := &fetchMetrics{
		startTime: time.Now(),
	}

	for iterator.HasNext() {
		select {
		case <-ctx.Done():
			uc.logger.Info("fetch cancelled by context",
				"reason", ctx.Err(),
				"batches_processed", metrics.batchesProcessed,
				"total_saved", metrics.totalSaved)
			return ctx.Err()
		default:
		}

		if err := uc.processBatch(ctx, iterator, cutoffDate, metrics); err != nil {
			if errors.Is(err, errCutoffReached) {
				uc.logger.Info("cutoff date reached, stopping fetch",
					"batches_processed", metrics.batchesProcessed,
					"total_saved", metrics.totalSaved,
					"total_failed", metrics.totalFailed)
				break
			}
			return err
		}

		time.Sleep(uc.rateLimitDelay)
	}

	metrics.duration = time.Since(metrics.startTime)
	uc.logFinalMetrics(metrics)

	return nil
}

var errCutoffReached = errors.New("cutoff date reached")

func (uc *FetchCarNumberUseCase) processBatch(
	ctx context.Context,
	iterator domain.NumberIterator,
	cutoffDate time.Time,
	metrics *fetchMetrics,
) error {
	metrics.batchesProcessed++

	uc.logger.Debug("fetching batch", "batch_number", metrics.batchesProcessed)

	numbers, err := iterator.Next(ctx)
	if err != nil {
		uc.logger.Error("failed to fetch batch",
			"batch_number", metrics.batchesProcessed,
			"error", err)
		return fmt.Errorf("fetch batch %d: %w", metrics.batchesProcessed, err)
	}

	if len(numbers) == 0 {
		uc.logger.Info("received empty batch, stopping",
			"batch_number", metrics.batchesProcessed)
		return errCutoffReached
	}

	uc.logger.Info("fetched batch",
		"batch_number", metrics.batchesProcessed,
		"batch_size", len(numbers))

	saved := 0
	for _, num := range numbers {
		if _, err := uc.repository.UpdateOrCreate(ctx, &num); err != nil {
			uc.logger.Warn("failed to save car number",
				"number", num.Number,
				"price", num.Price,
				"posted_at", num.PostedAt,
				"error", err)
			metrics.totalFailed++
			continue
		}
		saved++
	}

	metrics.totalSaved += saved

	uc.logger.Info("processed batch",
		"batch_number", metrics.batchesProcessed,
		"saved", saved,
		"failed", len(numbers)-saved,
		"total_saved", metrics.totalSaved,
		"total_failed", metrics.totalFailed)

	lastNumber := numbers[len(numbers)-1]
	if lastNumber.PostedAt.Before(cutoffDate) {
		uc.logger.Info("last number before cutoff date",
			"last_number", lastNumber.Number,
			"last_date", lastNumber.PostedAt,
			"cutoff_date", cutoffDate)
		return errCutoffReached
	}

	return nil
}

func (uc *FetchCarNumberUseCase) logFinalMetrics(metrics *fetchMetrics) {
	uc.logger.Info("fetch completed",
		"batches_processed", metrics.batchesProcessed,
		"total_saved", metrics.totalSaved,
		"total_failed", metrics.totalFailed,
		"duration_seconds", metrics.duration.Seconds(),
		"avg_batch_time_ms", metrics.avgBatchTime().Milliseconds())
}

type fetchMetrics struct {
	startTime        time.Time
	duration         time.Duration
	batchesProcessed int
	totalSaved       int
	totalFailed      int
}

func (m *fetchMetrics) avgBatchTime() time.Duration {
	if m.batchesProcessed == 0 {
		return 0
	}
	return m.duration / time.Duration(m.batchesProcessed)
}

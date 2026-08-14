package postgres

import (
	"context"
	"data-service/internal/domain"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

// stuckJobTimeout - через сколько взятая джоба считается брошенной и уходит в повторный прогон
const stuckJobTimeout = 2 * time.Hour

// JobRepository - хранение задач
type JobRepository struct {
	postgres *Postgres
}

func NewJobRepository(postgres *Postgres) *JobRepository {
	return &JobRepository{postgres: postgres}
}

const getJobQuery = `
SELECT id, name, start_after, payload, COALESCE(error, '')
FROM jobs
WHERE queue = $1
  AND start_after <= NOW()
  AND (status = $2 OR (status = $3 AND locked_at <= $4))
ORDER BY start_after
LIMIT 1
FOR UPDATE SKIP LOCKED
`

const lockJobQuery = `
UPDATE jobs
SET status = $1, locked_at = NOW(), updated_at = NOW()
WHERE id = $2
RETURNING locked_at
`

const failJobQuery = `
UPDATE jobs
SET status = $1, locked_at = NULL, error = $2, updated_at = NOW()
WHERE id = $3
`

const deleteJobQuery = `
DELETE FROM jobs
WHERE id = $1
`

const createJobQuery = `
INSERT INTO jobs (id, name, queue, status, start_after, payload) VALUES ($1, $2, $3, $4, $5, $6);`

func (r *JobRepository) GetJob(ctx context.Context, queue domain.JobQueue) (*domain.Job, error) {
	tx, err := r.postgres.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var (
		id         uuid.UUID
		name       string
		startAfter time.Time
		payload    string
		errorText  string
	)

	err = tx.QueryRow(ctx, getJobQuery,
		string(queue),
		string(domain.JobStatusPending),
		string(domain.JobStatusRunning),
		time.Now().Add(-stuckJobTimeout),
	).Scan(&id, &name, &startAfter, &payload, &errorText)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNoJob
	}
	if err != nil {
		return nil, fmt.Errorf("failed scan job: %w", err)
	}

	var lockedAt time.Time

	err = tx.QueryRow(ctx, lockJobQuery, string(domain.JobStatusRunning), id).Scan(&lockedAt)
	if err != nil {
		return nil, fmt.Errorf("failed lock job: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	job, err := domain.RestoreJob(
		id,
		domain.JobName(name),
		queue,
		domain.JobStatusRunning,
		startAfter,
		payload,
		&lockedAt,
		errorText,
	)
	if err != nil {
		return nil, fmt.Errorf("failed restore job: %w", err)
	}

	return job, nil
}

func (r *JobRepository) MarkJobFailed(ctx context.Context, id uuid.UUID, jobErr error) error {
	var errorText string
	if jobErr != nil {
		errorText = jobErr.Error()
	}

	_, err := r.postgres.pool.Exec(ctx, failJobQuery, string(domain.JobStatusFailed), errorText, id)
	if err != nil {
		return fmt.Errorf("failed mark job failed: %w", err)
	}

	return nil
}

func (r *JobRepository) DeleteJob(ctx context.Context, id uuid.UUID) error {
	_, err := r.postgres.pool.Exec(ctx, deleteJobQuery, id)
	if err != nil {
		return fmt.Errorf("failed delete job: %w", err)
	}

	return nil
}

func (r *JobRepository) CreateJob(ctx context.Context, job domain.Job) (*domain.Job, error) {
	_, err := r.postgres.pool.Exec(ctx, createJobQuery, job.Id, string(job.Name), string(job.Queue), string(job.Status), job.StartAfter, job.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed create job: %w", err)
	}

	return &job, nil
}

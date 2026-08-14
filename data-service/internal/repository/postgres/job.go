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

// JobRepository - хранение задач
type JobRepository struct {
	postgres *Postgres
}

func NewJobRepository(postgres *Postgres) *JobRepository {
	return &JobRepository{postgres: postgres}
}

const getJobQuery = `
SELECT id, name, queue, status, start_after, payload
FROM jobs
WHERE status = $1 AND start_after <= NOW()
ORDER BY start_after
LIMIT 1
FOR UPDATE SKIP LOCKED
`

const updateJobStatusQuery = `
UPDATE jobs
SET status = $1
WHERE id = $2
`
const createJobQuery = `
INSERT INTO jobs (id, name, queue, status, start_after, payload) VALUES ($1, $2, $3, $4, $5, $6);`

func (r *JobRepository) GetJob(ctx context.Context) (*domain.Job, error) {
	tx, err := r.postgres.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var (
		id         uuid.UUID
		name       string
		queue      string
		status     string
		startAfter time.Time
		payload    string
	)

	err = tx.QueryRow(ctx, getJobQuery, string(domain.JobStatusWaiting)).
		Scan(&id, &name, &queue, &status, &startAfter, &payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNoJob
	}
	if err != nil {
		return nil, fmt.Errorf("failed scan job: %w", err)
	}

	status = string(domain.JobStatusPending)

	_, err = tx.Exec(ctx, updateJobStatusQuery, status, id)
	if err != nil {
		return nil, fmt.Errorf("failed update job status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	job, err := domain.RestoreJob(id, domain.JobName(name), domain.JobQueue(queue), domain.JobStatus(status), startAfter, payload)
	if err != nil {
		return nil, fmt.Errorf("failed restore job: %w", err)
	}

	return job, nil
}

func (r *JobRepository) UpdateJobStatus(ctx context.Context, id uuid.UUID, status domain.JobStatus) error {
	_, err := r.postgres.pool.Exec(ctx, updateJobStatusQuery, string(status), id)
	if err != nil {
		return fmt.Errorf("failed update job status: %w", err)
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

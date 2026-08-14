package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNoJob - в очереди нет задач, готовых к исполнению
var ErrNoJob = errors.New("no job available")

type JobName string

const (
	JobNameImportAutonomerOffers JobName = "import-worker-offers"
)

type JobStatus string

const (
	JobStatusWaiting JobStatus = "waiting"
	JobStatusPending JobStatus = "pending"
	JobStatusFailed  JobStatus = "failed"
	JobStatusSuccess JobStatus = "success"
)

type JobQueue string

const (
	JobQueueDefault JobQueue = "default"
)

type Job struct {
	Id         uuid.UUID `validate:"required"`      // Id - Идентификатор
	Name       JobName   `validate:"required"`      // Name - Название джобы
	Queue      JobQueue  `validate:"required"`      // Queue - Очередь
	Status     JobStatus `validate:"required"`      // JobStatus - Статус исполнения
	StartAfter time.Time `validate:"required"`      // StartAfter - Отложенный запуск
	Payload    string    `validate:"required,json"` // Payload - Json с входными данными
}

func RestoreJob(
	id uuid.UUID,
	name JobName,
	queue JobQueue,
	status JobStatus,
	startAfter time.Time,
	payload string,
) (*Job, error) {
	return newJob(id, name, queue, status, startAfter, payload)
}

func NewJob(
	name JobName,
	queue JobQueue,
	status JobStatus,
	startAfter time.Time,
	payload string,
) (*Job, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}

	return newJob(id, name, queue, status, startAfter, payload)
}

func newJob(
	id uuid.UUID,
	name JobName,
	queue JobQueue,
	status JobStatus,
	startAfter time.Time,
	payload string,
) (*Job, error) {
	j := &Job{
		Id:         id,
		Name:       name,
		Queue:      queue,
		Status:     status,
		StartAfter: startAfter,
		Payload:    payload,
	}

	if err := validate.Struct(j); err != nil {
		return nil, err
	}

	return j, nil
}

package job

import (
	"data-service/internal/domain"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type task struct {
	name domain.JobName `validate:"required"`

	queue domain.JobQueue `validate:"required"`

	payload string `validate:"required"`

	uniqueKey string `validate:"required"`
}

func newTask(
	name domain.JobName,
	queue domain.JobQueue,
	payload string,
	uniqueKey string,
) (*task, error) {
	t := task{
		name:      name,
		queue:     queue,
		payload:   payload,
		uniqueKey: uniqueKey,
	}

	if err := validator.New().Struct(t); err != nil {
		return nil, fmt.Errorf("task create error: ", err)
	}

	return &t, nil
}

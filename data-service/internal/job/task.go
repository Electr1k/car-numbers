package job

import (
	"data-service/internal/domain"
)

// task - описание задачи, которую продюсер ставит в очередь
type task struct {
	name domain.JobName

	queue domain.JobQueue

	payload string

	uniqueKey string
}

func newTask(
	name domain.JobName,
	queue domain.JobQueue,
	payload string,
	uniqueKey string,
) *task {
	return &task{
		name:      name,
		queue:     queue,
		payload:   payload,
		uniqueKey: uniqueKey,
	}
}

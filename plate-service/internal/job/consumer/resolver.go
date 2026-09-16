package consumer

import (
	"context"
	"fmt"
	"plate-service/internal/domain"
)

type Handler interface {
	Handle(ctx context.Context, payload string) error
}

type Resolver struct {
	handlers map[domain.JobName]Handler
}

func NewResolver() *Resolver {
	return &Resolver{handlers: make(map[domain.JobName]Handler)}
}

func (r *Resolver) Register(name domain.JobName, handler Handler) {
	r.handlers[name] = handler
}

func (r *Resolver) Resolve(name domain.JobName) (Handler, error) {
	handler, ok := r.handlers[name]
	if !ok {
		return nil, fmt.Errorf("unknown job name: %s", name)
	}

	return handler, nil
}

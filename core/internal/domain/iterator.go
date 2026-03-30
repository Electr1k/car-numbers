package domain

import "context"

// NumberIterator defines the interface for iterating over car numbers
type NumberIterator interface {
	Next(ctx context.Context) ([]CarNumber, error)
	HasNext() bool
}

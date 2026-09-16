package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// newID - Идентификатор доменной сущности
//
// UUIDv7 (RFC 9562): старшие 48 бит - миллисекундный таймстемп, поэтому
// ключи монотонно возрастают и вставка остаётся локальной в конце B-tree,
// в отличие от случайного v4
func newID() (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("generate id: %w", err)
	}

	return id, nil
}

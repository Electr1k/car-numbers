package feed

import (
	"data-service/internal/domain/data"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

func encodeCursor(cursor data.FeedCursor) (string, error) {
	raw, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func decodeCursor(encoded string) (data.FeedCursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return data.FeedCursor{}, fmt.Errorf("decode cursor: %w", err)
	}

	var cursor data.FeedCursor
	if err := json.Unmarshal(raw, &cursor); err != nil {
		return data.FeedCursor{}, fmt.Errorf("decode cursor: %w", err)
	}
	if cursor.RefreshedAt.IsZero() || cursor.UpdatedAt.IsZero() || cursor.ID == uuid.Nil {
		return data.FeedCursor{}, errors.New("decode cursor: incomplete cursor")
	}

	return cursor, nil
}

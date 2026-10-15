package profile

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"plate-service/internal/domain/data"

	"github.com/google/uuid"
)

func encodeCursor(cursor data.ProfileOffersCursor) (string, error) {
	raw, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func decodeCursor(encoded string) (data.ProfileOffersCursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return data.ProfileOffersCursor{}, fmt.Errorf("decode cursor: %w", err)
	}

	var cursor data.ProfileOffersCursor
	if err := json.Unmarshal(raw, &cursor); err != nil {
		return data.ProfileOffersCursor{}, fmt.Errorf("decode cursor: %w", err)
	}
	if cursor.ID == uuid.Nil || cursor.RefreshedAt.IsZero() {
		return data.ProfileOffersCursor{}, errors.New("decode cursor: incomplete cursor")
	}

	return cursor, nil
}

package data

import (
	"data-service/internal/domain"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FeedNumber - номер для выдачи в свежих номерах
type FeedNumber struct {
	ID              uuid.UUID         // ID - Идентификатор
	Number          string            // Number - Номер
	RegionName      *string           // RegionName - Название региона
	RegionCode      *string           // RegionCode - Код региона
	Price           *float64          // Price - Минимальная цена
	Type            domain.NumberType // Type - Тип номера
	Count           int               // Count - количество офферов
	RefreshedAt     time.Time         // RefreshedAt - максимальная дата обновления оффера у провайдера
	UpdatedAt       time.Time         // UpdatedAt - максимальная дата обновления оффера в системе
	ReissueIncluded *bool             // ReissueIncluded - включено ли переоформление
}

type FeedCursor struct {
	RefreshedAt time.Time
	UpdatedAt   time.Time
	ID          uuid.UUID
}

func EncodeFeedCursor(refreshedAt time.Time, UpdatedAt time.Time, id uuid.UUID) string {
	payload := fmt.Sprintf("%d_%d_%s", refreshedAt.Unix(), UpdatedAt.Unix(), id)
	return base64.URLEncoding.EncodeToString([]byte(payload))
}

func DecodeFeedCursor(encode string) (*FeedCursor, error) {

	decode, err := base64.URLEncoding.DecodeString(encode)
	if err != nil {
		return nil, err
	}

	parts := strings.Split(string(decode), "_")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid feed cursor: %s", decode)
	}

	refreshedAt, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}
	updatedAt, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	return &FeedCursor{
		RefreshedAt: time.Unix(int64(refreshedAt), 0),
		UpdatedAt:   time.Unix(int64(updatedAt), 0),
		ID:          uuid.MustParse(parts[2]),
	}, nil
}

func GetEmptyFeedCursor() FeedCursor {
	return FeedCursor{
		RefreshedAt: time.Now(),
		UpdatedAt:   time.Now(),
		ID:          uuid.New(),
	}
}

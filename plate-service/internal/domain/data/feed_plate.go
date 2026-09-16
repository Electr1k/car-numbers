package data

import (
	"plate-service/internal/domain"
	"time"

	"github.com/google/uuid"
)

// PlateSort - порядок выдачи номеров
type PlateSort string

const (
	PlateSortUpdatedDesc PlateSort = "updated_desc"
	PlateSortPriceAsc    PlateSort = "price_asc"
	PlateSortPriceDesc   PlateSort = "price_desc"
)

// Valid - известен ли порядок выдачи
func (s PlateSort) Valid() bool {
	switch s {
	case PlateSortUpdatedDesc, PlateSortPriceAsc, PlateSortPriceDesc:
		return true
	}
	return false
}

// FeedPlate - номер для выдачи в свежих номерах
type FeedPlate struct {
	ID              uuid.UUID        // ID - Идентификатор
	Number          string           // Number - Номер
	RegionID        *int             // RegionID - Идентификатор региона
	RegionName      *string          // RegionName - Название региона
	RegionCode      *string          // RegionCode - Код региона
	Price           *float64         // Price - Минимальная цена
	Type            domain.PlateType // Type - Тип номера
	Count           int              // Count - количество офферов
	RefreshedAt     time.Time        // RefreshedAt - максимальная дата обновления оффера у провайдера
	UpdatedAt       time.Time        // UpdatedAt - максимальная дата обновления оффера в системе
	ReissueIncluded *bool            // ReissueIncluded - включено ли переоформление
}

// FeedCursor - позиция в выдаче номеров для сортировки Sort
type FeedCursor struct {
	Sort        PlateSort  `json:"s"`
	ID          uuid.UUID  `json:"id"`
	RefreshedAt *time.Time `json:"r,omitempty"`
	UpdatedAt   *time.Time `json:"u,omitempty"`
	Price       *string    `json:"p,omitempty"`
}

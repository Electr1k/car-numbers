package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// FeatureKey - Ключи фичи
type FeatureKey string

const (
	// FeatureKeyImportAutonomeraOffers - Импорт предложений из autonomera777
	FeatureKeyImportAutonomeraOffers FeatureKey = "import-autonomera-offers"

	// FeatureKeyDispatchImportAutonomeraOffers - Вызов импорта предложений из autonomera777
	FeatureKeyDispatchImportAutonomeraOffers FeatureKey = "dispatch-import-autonomera-offers"

	// FeatureKeyImportOfferDetail - Импорт детальной информации об оффере (из любого провайдера)
	FeatureKeyImportOfferDetail FeatureKey = "import-offer-detail"

	// FeatureKeyDispatchImportOfferDetail - Вызов импорта детальной информации об оффере (из любого провайдера)
	FeatureKeyDispatchImportOfferDetail FeatureKey = "dispatch-import-offer-detail"

	// FeatureKeyImportGosnomeruOffers - Импорт предложений из gosnomeru
	FeatureKeyImportGosnomeruOffers FeatureKey = "import-gosnomeru-offers"
)

var (
	ErrFeatureNotFound = errors.New("feature not found")
)

// Feature - Фичи (функционал)
type Feature struct {
	Id        uuid.UUID  `validate:"required"` // Id - Идентификатор
	Key       FeatureKey `validate:"required"` // Key - Ключ фичи
	Name      string     `validate:"required"` // Name - Название фичи
	Active    bool       // Active - Флаг активности фичи
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// RestoreFeature - Восстановление существующего фичи из хранилища
func RestoreFeature(
	id uuid.UUID,
	key FeatureKey,
	name string,
	active bool,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Feature, error) {
	return newFeature(id, key, name, active, createdAt, updatedAt)
}

func newFeature(
	id uuid.UUID,
	key FeatureKey,
	name string,
	active bool,
	createdAt *time.Time,
	updatedAt *time.Time,
) (*Feature, error) {
	f := &Feature{
		Id:        id,
		Key:       key,
		Name:      name,
		Active:    active,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if err := validate.Struct(f); err != nil {
		return nil, err
	}

	return f, nil
}

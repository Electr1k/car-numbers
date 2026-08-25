package feature

import (
	"context"
	"data-service/internal/domain"
	"errors"
	"fmt"
	"log/slog"
)

type Storage interface {
	GetFeatureByKey(ctx context.Context, key domain.FeatureKey) (*domain.Feature, error)
}

// Feature - проверка фича-флага перед запуском операции
type Feature struct {
	storage Storage
	logger  *slog.Logger
}

func NewFeature(storage Storage, logger *slog.Logger) *Feature {
	return &Feature{
		storage: storage,
		logger:  logger,
	}
}

// Enabled - включена ли фича
func (g *Feature) Enabled(ctx context.Context, key domain.FeatureKey) (bool, error) {
	feature, err := g.storage.GetFeatureByKey(ctx, key)
	if errors.Is(err, domain.ErrFeatureNotFound) {
		g.logger.Warn("feature not found, treating as disabled", "feature", key)
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get feature %s: %w", key, err)
	}

	if !feature.Active {
		g.logger.Info("feature disabled, skipping", "feature", key)
	}

	return feature.Active, nil
}

package postgres

import (
	"context"
	"data-service/internal/domain"
	"fmt"
)

// RegionRepository - хранение регионов и кодов
type RegionRepository struct {
	postgres *Postgres
}

func NewRegionRepository(postgres *Postgres) *RegionRepository {
	return &RegionRepository{postgres: postgres}
}

// getRegions - список регионов
const getRegions = `
SELECT id, name, code FROM regions
JOIN region_codes ON regions.id = region_codes.region_id
ORDER BY id ASC;`

// GetRegions - Возвращает регионы с их кодами
func (r *RegionRepository) GetRegions(ctx context.Context) ([]domain.RegionWithCodes, error) {
	rows, err := r.postgres.pool.Query(ctx, getRegions)
	if err != nil {
		return nil, fmt.Errorf("get regions: %w", err)
	}
	defer rows.Close()

	regions := make([]domain.RegionWithCodes, 0)
	indexByRegionID := make(map[int]int)
	for rows.Next() {
		var (
			id   int
			name string
			code string
		)

		if err := rows.Scan(&id, &name, &code); err != nil {
			return nil, fmt.Errorf("get region row: %w", err)
		}

		domainCode, err := domain.RestoreRegionCode(code, id)
		if err != nil {
			return nil, fmt.Errorf("get region row: %w", err)
		}

		if i, ok := indexByRegionID[id]; ok {
			regions[i].RegionCodes = append(regions[i].RegionCodes, *domainCode)
			continue
		}

		domainRegion, err := domain.RestoreRegion(id, name)
		if err != nil {
			return nil, fmt.Errorf("get region row: %w", err)
		}

		indexByRegionID[id] = len(regions)
		regions = append(regions, domain.RegionWithCodes{
			Region:      *domainRegion,
			RegionCodes: []domain.RegionCode{*domainCode},
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get regions: %w", err)
	}

	return regions, nil
}

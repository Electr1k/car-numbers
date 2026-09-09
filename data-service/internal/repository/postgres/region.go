package postgres

import (
	"context"
	"data-service/internal/domain"
	"fmt"
	"maps"
	"slices"
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

	codesByRegionsID := make(map[int]domain.RegionWithCodes)
	for rows.Next() {
		var (
			id   int
			name string
			code string
		)

		err = rows.Scan(&id, &name, &code)
		if err != nil {
			return nil, fmt.Errorf("get region row: %w", err)
		}

		domainCode, err := domain.RestoreRegionCode(code, id)
		if err != nil {
			return nil, fmt.Errorf("get region row: %w", err)
		}

		if region, ok := codesByRegionsID[id]; ok {
			region.RegionCodes = append(region.RegionCodes, *domainCode)
			codesByRegionsID[id] = region
		} else {
			domainRegion, err := domain.RestoreRegion(id, name)
			if err != nil {
				return nil, fmt.Errorf("get region row: %w", err)
			}

			codesByRegionsID[id] = domain.RegionWithCodes{
				Region:      *domainRegion,
				RegionCodes: []domain.RegionCode{*domainCode},
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get regions: %w", err)
	}

	return slices.Collect(maps.Values(codesByRegionsID)), nil
}

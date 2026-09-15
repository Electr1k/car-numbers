package region

import (
	"data-service/internal/domain"
	"slices"
)

type regionsResponse struct {
	Items []regionItem `json:"items"`
}

type regionItem struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Codes []string `json:"codes"`
}

// mapRegions - маппинг доменных регионов в ответ API
func mapRegions(regions []domain.RegionWithCodes) regionsResponse {
	items := make([]regionItem, 0, len(regions))
	for _, r := range regions {
		codes := make([]string, 0, len(r.RegionCodes))
		for _, c := range r.RegionCodes {
			codes = append(codes, c.Code)
		}
		slices.Sort(codes)

		items = append(items, regionItem{ID: r.Region.ID, Name: r.Region.Name, Codes: codes})
	}

	return regionsResponse{Items: items}
}

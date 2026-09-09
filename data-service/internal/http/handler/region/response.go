package region

import (
	"data-service/internal/domain"
	"sort"
)

type regionsResponse struct {
	Items []regionItem `json:"items"`
}

type regionItem struct {
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
		sort.Strings(codes)

		items = append(items, regionItem{Name: r.Region.Name, Codes: codes})
	}

	return regionsResponse{Items: items}
}

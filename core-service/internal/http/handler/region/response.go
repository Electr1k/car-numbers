package region

import (
	"core-service/internal/usecase/fetchregions"
)

type fetchRegionsResponse struct {
	Items []fetchRegionItem `json:"items"`
}

type fetchRegionItem struct {
	ID    int      `json:"id"`
	Name  string   `json:"name"`
	Codes []string `json:"codes"`
}

func mapFetchRegionsResponse(result fetchregions.Result) fetchRegionsResponse {
	items := make([]fetchRegionItem, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, fetchRegionItem{
			ID:    item.ID,
			Name:  item.Name,
			Codes: item.Codes,
		})
	}

	return fetchRegionsResponse{
		Items: items,
	}
}

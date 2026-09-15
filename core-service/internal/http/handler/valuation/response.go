package valuation

import (
	"core-service/internal/usecase/valuation"
)

type valuationResponse struct {
	Number     string             `json:"number"`
	Region     *valuationRegion   `json:"region"`
	Price      valuationPrice     `json:"price"`
	Confidence string             `json:"confidence"`
	Breakdown  valuationBreakdown `json:"breakdown"`
}

type valuationRegion struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type valuationPrice struct {
	P25 int `json:"p25"`
	P50 int `json:"p50"`
	P75 int `json:"p75"`
}

type valuationBreakdown struct {
	Base  int                      `json:"base"`
	Items []valuationBreakdownItem `json:"items"`
}

type valuationBreakdownItem struct {
	Code       string  `json:"code"`
	Value      string  `json:"value"`
	Title      string  `json:"title"`
	Multiplier float64 `json:"multiplier"`
	Exact      bool    `json:"exact"`
}

func mapValuationResponse(result valuation.Result) valuationResponse {
	var region *valuationRegion
	if result.Region != nil {
		region = &valuationRegion{
			Code: result.Region.Code,
			Name: result.Region.Name,
		}
	}

	breakdownItems := make([]valuationBreakdownItem, 0, len(result.Breakdown.Items))
	for _, item := range result.Breakdown.Items {
		breakdownItems = append(breakdownItems, valuationBreakdownItem{
			Code:       item.Code,
			Value:      item.Value,
			Title:      item.Title,
			Multiplier: item.Multiplier,
			Exact:      item.Exact,
		})
	}

	return valuationResponse{
		Number: result.Number,
		Region: region,
		Price: valuationPrice{
			P25: result.Price.P25,
			P50: result.Price.P50,
			P75: result.Price.P75,
		},
		Confidence: result.Confidence,
		Breakdown: valuationBreakdown{
			Base:  result.Breakdown.Base,
			Items: breakdownItems,
		},
	}
}

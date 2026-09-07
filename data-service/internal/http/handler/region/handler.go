package region

import (
	"data-service/internal/domain"
	"data-service/internal/http/response"
	"data-service/internal/usecase/fetchregions"
	"log/slog"
	"net/http"
	"sort"
)

type Handler struct {
	uc     *fetchregions.UseCase
	logger *slog.Logger
}

func New(uc *fetchregions.UseCase, logger *slog.Logger) *Handler {
	return &Handler{
		uc:     uc,
		logger: logger,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	res, err := h.uc.Handle(r.Context())
	if err != nil {
		h.logger.Error("fetch regions usecase failed", "error", err)
		response.WriteError(w, http.StatusBadGateway, "Произошла ошибка")
		return
	}

	response.WriteJSON(w, http.StatusOK, mapRegions(res))
}

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

package region

import (
	"data-service/internal/http/response"
	"data-service/internal/usecase/fetchregions"
	"log/slog"
	"net/http"
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
		response.WriteError(w, http.StatusBadGateway, "unexpected_error", "Произошла ошибка")
		return
	}

	response.WriteJSON(w, http.StatusOK, mapRegions(res))
}

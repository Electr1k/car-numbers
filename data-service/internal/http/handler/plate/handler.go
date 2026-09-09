package plate

import (
	"data-service/internal/domain"
	"data-service/internal/http/response"
	"data-service/internal/usecase/fetchnumber"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	uc     *fetchnumber.UseCase
	logger *slog.Logger
}

func New(uc *fetchnumber.UseCase, logger *slog.Logger) *Handler {
	return &Handler{
		uc:     uc,
		logger: logger,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	plateId, err := uuid.Parse(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "validation_error", "Некорректные параметры запроса")
		return
	}

	res, err := h.uc.Handle(r.Context(), plateId)
	if errors.Is(err, domain.ErrNumberNotFound) {
		response.WriteError(w, http.StatusNotFound, "not_found", "Номер не найден")
		return
	}
	if err != nil {
		h.logger.Error("fetch plate usecase failed", "error", err)
		response.WriteError(w, http.StatusBadGateway, "unexpected_error", "Произошла ошибка")
		return
	}

	response.WriteJSON(w, http.StatusOK, mapNumber(*res))
}

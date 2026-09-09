package feed

import (
	"data-service/internal/http/request"
	"data-service/internal/http/response"
	"data-service/internal/usecase/fetchfeednumbers"
	"log/slog"
	"net/http"
)

type Handler struct {
	uc     *fetchfeednumbers.UseCase
	logger *slog.Logger
}

func New(uc *fetchfeednumbers.UseCase, logger *slog.Logger) *Handler {
	return &Handler{
		uc:     uc,
		logger: logger,
	}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	req := newFeedRequest()
	if err := request.DecodeQuery(r, &req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "Некорректные параметры запроса")
		return
	}

	if err := request.Validate(req); err != nil {
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := h.uc.Handle(r.Context(), fetchfeednumbers.Params{Limit: req.Limit, Offset: req.Offset})
	if err != nil {
		h.logger.Error("fetch feed offers usecase failed", "error", err)
		response.WriteError(w, http.StatusBadGateway, "Произошла ошибка")
		return
	}

	response.WriteJSON(w, http.StatusOK, mapFeedNumbers(res))
}

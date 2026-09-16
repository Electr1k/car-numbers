package valuation

import (
	"core-service/internal/domain"
	"core-service/internal/service"
)

type valuationRequest struct {
	Number string `form:"number"`
}

func newValuationRequest() valuationRequest {
	return valuationRequest{}
}

func validateNumber(number string) error {
	switch {
	case number == "":
		return service.NewClientError(service.ErrBadRequest, "number_required", "Укажите номер для оценки.")
	case domain.IsMotoPlate(number):
		return service.NewClientError(service.ErrUnprocessable, "moto", "Оценка возможна только для автономеров.")
	case domain.HasPlateMask(number):
		return service.NewClientError(service.ErrUnprocessable, "mask_not_supported", "Неверный формат номера.")
	case !domain.IsCanonicalPlate(number):
		return service.NewClientError(service.ErrUnprocessable, "format", "Номер вводится как А123ВС777.")
	}

	return nil
}

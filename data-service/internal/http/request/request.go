package request

import (
	"net/http"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
)

var (
	decoder  = form.NewDecoder()
	validate = validator.New(validator.WithRequiredStructEnabled())
)

// DecodeQuery - разбирает query-параметры в структуру
func DecodeQuery(r *http.Request, object any) error {
	return decoder.Decode(object, r.URL.Query())
}

func Validate(v any) error {
	return validate.Struct(v)
}

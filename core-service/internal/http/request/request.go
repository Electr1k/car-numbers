package request

import (
	"net/http"

	"github.com/go-playground/form/v4"
)

var decoder = form.NewDecoder()

// DecodeQuery - разбирает query-параметры в структуру
func DecodeQuery(r *http.Request, object any) error {
	return decoder.Decode(object, r.URL.Query())
}

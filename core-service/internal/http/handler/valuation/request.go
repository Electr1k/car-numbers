package valuation

import (
	"regexp"
	"strings"
)

type valuationRequest struct {
	Number string `form:"number"`
}

func newValuationRequest() valuationRequest {
	return valuationRequest{}
}

var plateMask = regexp.MustCompile(
	`^[АВЕКМНОРСТУХ]\d{3}[АВЕКМНОРСТУХ]{2}\d{2,3}$`,
)

var latinToCyrillic = map[rune]rune{
	'A': 'А', 'B': 'В', 'E': 'Е', 'K': 'К', 'M': 'М',
	'H': 'Н', 'O': 'О', 'P': 'Р', 'C': 'С', 'T': 'Т',
	'Y': 'У', 'X': 'Х',
}

// normalizeNumber нормализует номер
func normalizeNumber(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")

	var b strings.Builder
	for _, r := range s {
		if c, ok := latinToCyrillic[r]; ok {
			r = c
		}
		b.WriteRune(r)
	}
	return b.String()
}

// isValidNumber валидация
func isValidNumber(s string) bool {
	return plateMask.MatchString(s)
}

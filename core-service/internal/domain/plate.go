package domain

import (
	"regexp"
	"strings"
)

const PlateMaskChar = "*"

var (
	canonicalPlate = regexp.MustCompile(`^[АВЕКМНОРСТУХ][0-9]{3}[АВЕКМНОРСТУХ]{2}[0-9]{2,3}$`)
	motoPlate      = regexp.MustCompile(`^[0-9]{4}[АВЕКМНОРСТУХ]{2}[0-9]{2,3}$`)
)

var latinToCyrillic = map[rune]rune{
	'A': 'А', 'B': 'В', 'E': 'Е', 'K': 'К', 'M': 'М',
	'H': 'Н', 'O': 'О', 'P': 'Р', 'C': 'С', 'T': 'Т',
	'Y': 'У', 'X': 'Х',
}

func NormalizePlate(s string) string {
	s = strings.ToUpper(strings.Join(strings.Fields(s), ""))

	return strings.Map(func(r rune) rune {
		if c, ok := latinToCyrillic[r]; ok {
			return c
		}
		return r
	}, s)
}

func IsCanonicalPlate(s string) bool {
	return canonicalPlate.MatchString(s)
}

func IsMotoPlate(s string) bool {
	return motoPlate.MatchString(s)
}

func HasPlateMask(s string) bool {
	return strings.Contains(s, PlateMaskChar)
}

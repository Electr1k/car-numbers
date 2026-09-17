package domain

import (
	"strconv"
	"strings"
)

const (
	plateMask = "*"

	plateAlphabet = "АВЕКМНОРСТУХ"

	digitsAscending = "0123456789"
)

// plateParts - разобранный номер на части
type plateParts struct {
	digits       string
	letters      string
	region       string
	digitsKnown  bool
	lettersKnown bool
	regionKnown  bool
}

// classifyPlate - классифицирует категорию номера
func classifyPlate(number string, plateType PlateType) []CategoryID {
	parts, ok := splitPlate(number, plateType)
	if !ok {
		return nil
	}

	categories := make([]CategoryID, 0, len(getCategoryIds()))
	for _, category := range getCategoryIds() {
		if parts.matches(category) {
			categories = append(categories, category)
		}
	}

	if len(categories) == 0 {
		return nil
	}

	return categories
}

// matches - подходит ли номер под категорию
func (p plateParts) matches(category CategoryID) bool {
	switch category {
	case CategoryIdSameDigits:
		return p.digitsKnown && allSame(p.digits)
	case CategoryIdFirstTen:
		return p.digitsKnown && strings.Trim(p.digits[:len(p.digits)-1], "0") == "" &&
			p.digits[len(p.digits)-1] != '0'
	case CategoryIdRoundHundreds:
		return p.digitsKnown && strings.HasSuffix(p.digits, "00") && !allZeros(p.digits)
	case CategoryIdZeroEdges:
		return p.digitsKnown && p.digits[0] == '0' && p.digits[len(p.digits)-1] == '0' && !allZeros(p.digits)
	case CategoryIdMirroredDigits:
		return p.digitsKnown && p.digits == reverse(p.digits) && !allSame(p.digits)
	case CategoryIdPairDigits:
		return p.digitsKnown && hasAdjacentPair(p.digits) && !allSame(p.digits)
	case CategoryIdDigitsStaircase:
		return p.digitsKnown && isStaircase(p.digits, digitsAscending)
	case CategoryIdDigitsAsRegion:
		return p.digitsKnown && p.regionKnown && sameNumber(p.digits, p.region)
	case CategoryIdSameLetters:
		return p.lettersKnown && allSame(p.letters)
	case CategoryIdMirroredLetters:
		return p.lettersKnown && hasSeries(p.letters) && p.letters == reverse(p.letters) && !allSame(p.letters)
	case CategoryIdPairLetters:
		return p.lettersKnown && hasSeries(p.letters) && hasAdjacentPair(p.letters) && !allSame(p.letters)
	case CategoryIdLetterStaircase:
		return p.lettersKnown && hasSeries(p.letters) && isStaircase(p.letters, plateAlphabet)
	}

	return false
}

// splitPlate - разбор номера на составляющие
func splitPlate(number string, plateType PlateType) (plateParts, bool) {
	runes := []rune(number)
	if len(runes) < 8 || len(runes) > 9 {
		return plateParts{}, false
	}

	var parts plateParts

	switch plateType {
	case PlateTypeCar:
		parts = plateParts{
			digits:  string(runes[1:4]),
			letters: string(runes[0:1]) + string(runes[4:6]),
			region:  string(runes[6:]),
		}
	case PlateTypeMoto:
		parts = plateParts{
			digits:  string(runes[0:4]),
			letters: string(runes[4:6]),
			region:  string(runes[6:]),
		}
	case PlateTypeTrailer:
		parts = plateParts{
			digits:  string(runes[2:6]),
			letters: string(runes[0:2]),
			region:  string(runes[6:]),
		}
	default:
		return plateParts{}, false
	}

	if !isDigitsOrMask(parts.digits) || !isLettersOrMask(parts.letters) || !isDigitsOrMask(parts.region) {
		return plateParts{}, false
	}

	parts.digitsKnown = !strings.Contains(parts.digits, plateMask)
	parts.lettersKnown = !strings.Contains(parts.letters, plateMask)
	parts.regionKnown = !strings.Contains(parts.region, plateMask)

	return parts, true
}

// hasSeries - есть ли у номера буквенная серия из трёх букв
func hasSeries(letters string) bool {
	return len([]rune(letters)) == 3
}

// allSame - все ли знаки одинаковые
func allSame(value string) bool {
	runes := []rune(value)
	for _, r := range runes {
		if r != runes[0] {
			return false
		}
	}

	return true
}

// allZeros - состоит ли значение из одних нулей
func allZeros(value string) bool {
	return strings.Trim(value, "0") == ""
}

// hasAdjacentPair - есть ли два одинаковых знака подряд
func hasAdjacentPair(value string) bool {
	runes := []rune(value)
	for i := 1; i < len(runes); i++ {
		if runes[i] == runes[i-1] {
			return true
		}
	}

	return false
}

// isStaircase - идут ли знаки подряд по алфавиту, вверх или вниз
func isStaircase(value, alphabet string) bool {
	return strings.Contains(alphabet, value) || strings.Contains(reverse(alphabet), value)
}

// sameNumber - равны ли цифры и код региона как числа
func sameNumber(digits, region string) bool {
	digitsValue, err := strconv.Atoi(digits)
	if err != nil {
		return false
	}

	regionValue, err := strconv.Atoi(region)
	if err != nil {
		return false
	}

	return digitsValue == regionValue
}

// reverse - разворот строки
func reverse(value string) string {
	runes := []rune(value)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

func isDigitsOrMask(value string) bool {
	for _, r := range value {
		if !strings.ContainsRune(digitsAscending+plateMask, r) {
			return false
		}
	}

	return true
}

func isLettersOrMask(value string) bool {
	for _, r := range value {
		if !strings.ContainsRune(plateAlphabet+plateMask, r) {
			return false
		}
	}

	return true
}

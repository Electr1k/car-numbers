package domain

import "slices"

type CategoryID string

const (
	// CategoryIdSameLetters - Одинаковые буквы
	CategoryIdSameLetters CategoryID = "same-letters"
	// CategoryIdMirroredLetters - Зеркальные буквы
	CategoryIdMirroredLetters CategoryID = "mirrored-letters"
	// CategoryIdPairLetters - Пара букв
	CategoryIdPairLetters CategoryID = "pair-letters"
	// CategoryIdLetterStaircase - Лесенка букв
	CategoryIdLetterStaircase CategoryID = "letter-staircase"
	// CategoryIdSameDigits - Одинаковые цифры
	CategoryIdSameDigits CategoryID = "same-digits"
	// CategoryIdFirstTen - Первая десятка
	CategoryIdFirstTen CategoryID = "first-ten"
	// CategoryIdRoundHundreds - Круглые сотни
	CategoryIdRoundHundreds CategoryID = "round-hundreds"
	// CategoryIdZeroEdges - Ноль по краям
	CategoryIdZeroEdges CategoryID = "zero-edges"
	// CategoryIdMirroredDigits - Зеркальные цифры
	CategoryIdMirroredDigits CategoryID = "mirrored-digits"
	// CategoryIdPairDigits - Пара цифр
	CategoryIdPairDigits CategoryID = "pair-digits"
	// CategoryIdDigitsStaircase - Лесенка цифр
	CategoryIdDigitsStaircase CategoryID = "digits-staircase"
	// CategoryIdDigitsAsRegion - Цифры как регион
	CategoryIdDigitsAsRegion CategoryID = "digits-as-region"
)

// GetCategoryIds Возвращает список категорий номера
func GetCategoryIds() []CategoryID {
	return []CategoryID{
		CategoryIdSameLetters,
		CategoryIdMirroredLetters,
		CategoryIdPairLetters,
		CategoryIdLetterStaircase,
		CategoryIdSameDigits,
		CategoryIdFirstTen,
		CategoryIdRoundHundreds,
		CategoryIdZeroEdges,
		CategoryIdMirroredDigits,
		CategoryIdPairDigits,
		CategoryIdDigitsStaircase,
		CategoryIdDigitsAsRegion,
	}
}

// Valid - известна ли категория
func (c CategoryID) Valid() bool {
	return slices.Contains(GetCategoryIds(), c)
}

// Category - Категория номера
type Category struct {
	ID   CategoryID `validate:"required"` // ID - Код категории
	Name string     `validate:"required"` // Name - Название категории
}

// RestoreCategory - Восстановление существующей категории из хранилища
func RestoreCategory(
	id CategoryID,
	name string,
) (*Category, error) {
	return newCategory(id, name)
}

func newCategory(
	id CategoryID,
	name string,
) (*Category, error) {
	n := &Category{
		ID:   id,
		Name: name,
	}

	if err := validate.Struct(n); err != nil {
		return nil, err
	}

	return n, nil
}

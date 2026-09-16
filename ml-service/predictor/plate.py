"""Разбор российского ГРЗ и категории номера"""

import re
from dataclasses import dataclass
from enum import StrEnum
from itertools import product

# Буквы, допустимые в ГРЗ: те, что совпадают по начертанию с латиницей
LETTERS = 'АВЕКМНОРСТУХ'

# Символ-заглушка вместо неуказанного знака в объявлении
MASK = '*'

# Канонический формат: буква, три цифры, две буквы, регион в 2-3 цифры
CANONICAL_PATTERN = rf'[{LETTERS}][0-9]{{3}}[{LETTERS}]{{2}}[0-9]{{2,3}}'

# Тот же формат, но с допустимой маской в любой позиции
MASKED_PATTERN = (
    rf'[{LETTERS}\{MASK}][0-9\{MASK}]{{3}}[{LETTERS}\{MASK}]{{2}}[0-9\{MASK}]{{2,3}}'
)

# Мото и прицепы: четыре цифры и две буквы. Отдельно — чтобы внятно отказать
MOTO_PATTERN = rf'[0-9]{{4}}[{LETTERS}]{{2}}[0-9]{{2,3}}'

_CANONICAL = re.compile(CANONICAL_PATTERN)
_MASKED = re.compile(MASKED_PATTERN)
_MOTO = re.compile(MOTO_PATTERN)
_MOTO_MASKED = re.compile(rf'[0-9\{MASK}]{{4}}[{LETTERS}\{MASK}]{{2}}[0-9\{MASK}]{{2,3}}')

_DIGITS = '0123456789'

# Латиница-двойник → кириллица: раскладка у пользователя может быть любой
_LATIN_LOOKALIKES = str.maketrans('ABEKMHOPCTYX', LETTERS)

# Лесенки перечислены явно: 890 и 901 в них не входят
_LADDER_UP = frozenset({'012', '123', '234', '345', '456', '567', '678', '789'})
_LADDER_DOWN = frozenset({'210', '321', '432', '543', '654', '765', '876', '987'})


class DigitClass(StrEnum):
    """Класс паттерна цифровой части."""

    SAME = 'AAA'
    FIRST_TEN = '00X'
    ROUND_HUNDRED = 'X00'
    ZERO_EDGES = '0X0'
    MIRROR = 'ABA'
    PAIR_HEAD = 'AAB'
    PAIR_TAIL = 'ABB'
    LADDER_UP = 'LADDER_UP'
    LADDER_DOWN = 'LADDER_DOWN'
    OTHER = 'OTHER'


class LetterClass(StrEnum):
    """Класс буквенной серии для номеров с редкой серией."""

    SAME3 = 'SAME3'
    PAIR = 'PAIR'
    OTHER = 'OTHER'


DIGIT_CLASSES = tuple(DigitClass)
LETTER_CLASSES = tuple(LetterClass)


class PlateError(ValueError):
    """Номер не разбирается. `reason` — машинный код для ответа API."""

    FORMAT = 'format'
    MASKED = 'masked'
    MOTO = 'moto'

    def __init__(self, number: str, reason: str, message: str):
        super().__init__(message)
        self.number = number
        self.reason = reason


def digit_class(digits: str) -> DigitClass:
    """Класс цифровой части. Классы пересекаются — выигрывает первое совпадение."""
    a, b, c = digits
    if a == b == c:
        return DigitClass.SAME
    if a == '0' and b == '0':
        return DigitClass.FIRST_TEN
    if b == '0' and c == '0':
        return DigitClass.ROUND_HUNDRED
    if a == '0' and c == '0':
        return DigitClass.ZERO_EDGES
    if a == c:
        return DigitClass.MIRROR
    if a == b:
        return DigitClass.PAIR_HEAD
    if b == c:
        return DigitClass.PAIR_TAIL
    if digits in _LADDER_UP:
        return DigitClass.LADDER_UP
    if digits in _LADDER_DOWN:
        return DigitClass.LADDER_DOWN
    return DigitClass.OTHER


def letter_class(series: str) -> LetterClass:
    """Класс буквенной серии из трёх букв."""
    a, b, c = series
    if a == b == c:
        return LetterClass.SAME3
    if a == b or b == c or a == c:
        return LetterClass.PAIR
    return LetterClass.OTHER


def normalize(number: str) -> str:
    """Приведение ввода к каноническому виду: регистр, пробелы, латиница-двойники."""
    return number.strip().replace(' ', '').upper().translate(_LATIN_LOOKALIKES)


def mask_count(number: str) -> int:
    """Сколько знаков в номере не указано."""
    return normalize(number).count(MASK)


@dataclass(frozen=True, slots=True)
class Plate:
    """Разобранный номер. Поля — части ГРЗ, свойства — признаки модели."""

    number: str
    letter1: str
    digits: str
    letters23: str
    region: str

    @property
    def series(self) -> str:
        """Буквенная серия целиком: первая буква плюс пара."""
        return self.letter1 + self.letters23

    @property
    def digits_value(self) -> int:
        return int(self.digits)

    @property
    def region_value(self) -> int:
        return int(self.region)

    @property
    def digit_class(self) -> DigitClass:
        return digit_class(self.digits)

    @property
    def letter_class(self) -> LetterClass:
        return letter_class(self.series)

    @property
    def digits_eq_region(self) -> bool:
        """Числовое значение цифр совпадает с кодом региона: У001ОО01, К777ОС777."""
        return self.digits_value == self.region_value


def _reject(normalized: str) -> PlateError:
    """Отказ с точной причиной: мотоформат стоит назвать отдельно."""
    if _MOTO.fullmatch(normalized) or _MOTO_MASKED.fullmatch(normalized):
        return PlateError(
            normalized,
            PlateError.MOTO,
            f'номер {normalized} — мото или прицеп, такие мы не оцениваем',
        )
    return PlateError(
        normalized,
        PlateError.FORMAT,
        f'номер {normalized} не в формате Б999ББ + регион',
    )


def parse(number: str) -> Plate:
    """Разбор номера. Бросает PlateError, если формат не канонический."""
    normalized = normalize(number)
    if MASK in normalized:
        raise PlateError(
            normalized,
            PlateError.MASKED,
            f'номер {normalized} содержит неуказанные знаки',
        )
    if not _CANONICAL.fullmatch(normalized):
        raise _reject(normalized)
    return Plate(
        number=normalized,
        letter1=normalized[0],
        digits=normalized[1:4],
        letters23=normalized[4:6],
        region=normalized[6:],
    )


def alphabet_at(position: int) -> str:
    """Что может стоять в позиции канонического номера: буква или цифра."""
    return LETTERS if position in (0, 4, 5) else _DIGITS


def expand(number: str) -> list[Plate]:
    """Все допустимые подстановки вместо масок."""
    normalized = normalize(number)
    if not _MASKED.fullmatch(normalized):
        raise _reject(normalized)

    slots = [position for position, char in enumerate(normalized) if char == MASK]
    if not slots:
        return [parse(normalized)]

    variants = []
    for combination in product(*(alphabet_at(position) for position in slots)):
        chars = list(normalized)
        for position, char in zip(slots, combination):
            chars[position] = char
        variants.append(parse(''.join(chars)))
    return variants


def try_parse(number: str) -> Plate | None:
    """Разбор номера, который возвращает None вместо исключения."""
    try:
        return parse(number)
    except PlateError:
        return None

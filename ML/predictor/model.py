"""Артефакт модели: хранение, загрузка и предсказание"""

import json
import math
from dataclasses import dataclass, field
from pathlib import Path

from .features import RARE_REGION, plate_keys, plate_numeric
from .plate import Plate

MODEL_FILE = 'model.json'
CURRENT_FILE = 'current.json'

# Границы уровней доверия по предсказанной цене
HIGH_CONFIDENCE_FROM = 50_000
LOW_CONFIDENCE_FROM = 2_000_000

# Человекочитаемые названия групп — фронт выводит их как есть.
# Словари раздельные: OTHER у цифр и у букв должен читаться по-разному
_DIGIT_CLASS_TITLES = {
    'AAA': 'Три одинаковые цифры',
    '00X': 'Первая десятка',
    'X00': 'Круглые сотни',
    '0X0': 'Ноль по краям',
    'ABA': 'Зеркальные цифры',
    'AAB': 'Пара одинаковых цифр',
    'ABB': 'Пара одинаковых цифр',
    'LADDER_UP': 'Лесенка вверх',
    'LADDER_DOWN': 'Лесенка вниз',
    'OTHER': 'Обычные цифры',
}

_LETTER_CLASS_TITLES = {
    'SAME3': 'Три одинаковые буквы',
    'PAIR': 'Пара одинаковых букв',
    'OTHER': 'Разные буквы',
}


@dataclass(frozen=True)
class Item:
    """Вклад одной группы признаков: во сколько раз она меняет цену от базовой."""

    code: str
    value: str
    title: str
    multiplier: float
    exact: bool = True


@dataclass(frozen=True)
class Breakdown:
    """Раскладка цены. base, умноженная на все multiplier, даёт ровно p50."""

    base: int
    items: list[Item]


@dataclass(frozen=True)
class Estimate:
    """Оценка цены номера."""

    number: str
    p25: int
    p50: int
    p75: int
    confidence: str
    breakdown: Breakdown
    model_version: str


@dataclass(frozen=True)
class Model:
    """Обученная модель: коэффициенты и всё, что нужно для их применения."""

    version: str
    intercept: float
    weights: dict[str, float]
    numeric: dict[str, float]
    reference: float
    residuals: dict[str, float]
    centering: dict = field(default_factory=dict)
    metrics: dict = field(default_factory=dict)
    config: dict = field(default_factory=dict)
    vocabulary: dict = field(default_factory=dict)

    def log_price(self, plate: Plate) -> float:
        """Сумма коэффициентов: цена в логарифме, площадка нейтрализована."""
        total = self.intercept + self.reference
        for key in plate_keys(plate, self.weights):
            total += self.weights.get(key, 0.0)
        for name, value in plate_numeric(plate).items():
            total += self.numeric.get(name, 0.0) * value
        return total

    def estimate(self, plate: Plate) -> Estimate:
        center = self.log_price(plate)
        p50 = math.exp(center + self.residuals['p50'])
        return Estimate(
            number=plate.number,
            p25=round(math.exp(center + self.residuals['p25'])),
            p50=round(p50),
            p75=round(math.exp(center + self.residuals['p75'])),
            confidence=confidence_for(p50),
            breakdown=self.breakdown(plate),
            model_version=self.version,
        )

    def breakdown(self, plate: Plate) -> Breakdown:
        """Раскладка p50 на множители по группам признаков."""
        keys = {key.partition('=')[0]: key for key in plate_keys(plate, self.weights)}
        numeric = plate_numeric(plate)

        def logs(*names: str) -> float:
            """Вклад группы за вычетом её типичного уровня по рынку."""
            total = sum(self.weights.get(keys[n], 0.0) for n in names if n in keys)
            return total - sum(self.centering.get(n, 0.0) for n in names)

        series_key = keys.get('series', '')
        series_rare = series_key.startswith('series=RARE_')
        region_key = keys.get('reg', '')

        exact_region = region_key != RARE_REGION
        items = [
            Item('region', plate.region,
                 f'Регион {plate.region}' if exact_region else 'Редкий регион',
                 _times(logs('reg')), exact_region),
            # величина числа — про те же цифры, поэтому в один пункт с точным значением
            Item('digits', plate.digits, f'Цифры {plate.digits}',
                 _times(logs('dg', 'log_dnum')
                        + self.numeric.get('log_dnum', 0.0) * numeric['log_dnum'])),
            Item('digit_class', str(plate.digit_class),
                 _DIGIT_CLASS_TITLES[str(plate.digit_class)], _times(logs('dcls'))),
            Item('letter_class', str(plate.letter_class),
                 _LETTER_CLASS_TITLES[str(plate.letter_class)], _times(logs('scls'))),
            # взаимодействие с регионом — уточнение к той же серии
            Item('series', plate.series, f'Серия {plate.series}',
                 _times(logs('series', 'sxr')), not series_rare),
        ]
        if plate.digits_eq_region:
            items.append(
                Item('digits_eq_region', plate.region, 'Цифры совпадают с регионом',
                     _times(self.numeric.get('dg_eq_reg', 0.0)))
            )

        items.sort(key=lambda item: item.multiplier, reverse=True)
        # база включает типичные уровни всех групп, поэтому это цена обычного номера
        base = math.exp(
            self.intercept + self.reference + self.residuals['p50'] + sum(self.centering.values())
        )
        return Breakdown(base=round(base), items=items)

    # --- хранение ---

    def to_dict(self) -> dict:
        return {
            'version': self.version,
            'intercept': self.intercept,
            'reference': self.reference,
            'centering': self.centering,
            'residuals': self.residuals,
            'metrics': self.metrics,
            'config': self.config,
            'vocabulary': self.vocabulary,
            'numeric': self.numeric,
            'weights': self.weights,
        }

    @classmethod
    def from_dict(cls, payload: dict) -> 'Model':
        return cls(
            version=payload['version'],
            intercept=payload['intercept'],
            weights=payload['weights'],
            numeric=payload['numeric'],
            reference=payload['reference'],
            residuals=payload['residuals'],
            centering=payload.get('centering', {}),
            metrics=payload.get('metrics', {}),
            config=payload.get('config', {}),
            vocabulary=payload.get('vocabulary', {}),
        )


def _times(log_weight: float) -> float:
    """Логарифмический вклад — во сколько раз меняется цена."""
    return round(math.exp(log_weight), 4)


def confidence_for(price: float) -> str:
    """Уровень доверия по ценовому сегменту"""
    if price > LOW_CONFIDENCE_FROM:
        return 'low'
    if price < HIGH_CONFIDENCE_FROM:
        return 'medium'
    return 'high'


def _label(key: str) -> str:
    kind, _, value = key.partition('=')
    if kind == 'reg':
        return 'редкий регион' if value == 'RARE' else f'регион {value}'
    if kind == 'dg':
        return f'цифры {value}'
    if kind == 'series':
        if value.startswith('RARE_'):
            return _LETTER_CLASS_LABELS.get(value[5:], value)
        return f'серия {value}'
    if kind == 'dcls':
        return _DIGIT_CLASS_LABELS.get(value, value)
    if kind == 'scls':
        return _LETTER_CLASS_LABELS.get(value, value)
    if kind == 'sxr':
        series, _, region = value.partition('|')
        return f'серия {series} в регионе {region}'
    return key


def save(model: Model, models_dir: Path) -> Path:
    """Пишет неизменяемую версию артефакта. Активной её делает activate()."""
    directory = Path(models_dir) / model.version
    directory.mkdir(parents=True, exist_ok=True)
    path = directory / MODEL_FILE
    path.write_text(
        json.dumps(model.to_dict(), ensure_ascii=False, indent=1), encoding='utf-8'
    )
    return path


def activate(version: str, models_dir: Path) -> Path:
    path = Path(models_dir) / CURRENT_FILE
    path.write_text(json.dumps({'active': version}, ensure_ascii=False), encoding='utf-8')
    return path


def active_version(models_dir: Path) -> str | None:
    path = Path(models_dir) / CURRENT_FILE
    if not path.exists():
        return None
    return json.loads(path.read_text(encoding='utf-8'))['active']


def load_version(version: str, models_dir: Path) -> Model:
    path = Path(models_dir) / version / MODEL_FILE
    return Model.from_dict(json.loads(path.read_text(encoding='utf-8')))


def load_active(models_dir: Path) -> Model | None:
    version = active_version(models_dir)
    return load_version(version, models_dir) if version else None


class Registry:
    """Активная модель в памяти. Перечитывается, когда меняется current.json."""

    # Каталог версий: одинаков и в контейнере (WORKDIR /app), и при запуске из ML/
    MODELS_DIR = Path('models')

    def __init__(self, models_dir: Path | None = None):
        self.models_dir = Path(models_dir) if models_dir else self.MODELS_DIR
        self._model: Model | None = None
        self._stamp: tuple[str, float] | None = None

    def current(self) -> Model | None:
        path = self.models_dir / CURRENT_FILE
        if not path.exists():
            self._model, self._stamp = None, None
            return None

        stamp = (str(path), path.stat().st_mtime)
        if self._model is None or stamp != self._stamp:
            self._model = load_active(self.models_dir)
            self._stamp = stamp
        return self._model

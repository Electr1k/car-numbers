"""Артефакт модели: хранение, загрузка и предсказание"""

import json
import math
from dataclasses import asdict, dataclass, field
from pathlib import Path

from .features import plate_keys, plate_numeric
from .plate import Plate

MODEL_FILE = 'model.json'
CURRENT_FILE = 'current.json'

# Границы уровней доверия по предсказанной цене
HIGH_CONFIDENCE_FROM = 50_000
LOW_CONFIDENCE_FROM = 2_000_000

_DIGIT_CLASS_LABELS = {
    'AAA': 'три одинаковые цифры',
    '00X': 'первая десятка',
    'X00': 'круглые сотни',
    '0X0': 'ноль по краям',
    'ABA': 'зеркальные цифры',
    'AAB': 'пара цифр впереди',
    'ABB': 'пара цифр сзади',
    'LADDER_UP': 'лесенка вверх',
    'LADDER_DOWN': 'лесенка вниз',
    'OTHER': 'обычные цифры',
}

_LETTER_CLASS_LABELS = {
    'SAME3': 'три одинаковые буквы',
    'PAIR': 'пара одинаковых букв',
    'OTHER': 'разные буквы',
}


@dataclass(frozen=True)
class Factor:
    """Вклад одного признака в цену, выраженный множителем."""

    name: str
    weight: float


@dataclass(frozen=True)
class Estimate:
    """Оценка цены номера."""

    number: str
    p25: int
    p50: int
    p75: int
    confidence: str
    factors: list[Factor]
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

    def estimate(self, plate: Plate, factor_limit: int = 4) -> Estimate:
        center = self.log_price(plate)
        p50 = math.exp(center + self.residuals['p50'])
        return Estimate(
            number=plate.number,
            p25=round(math.exp(center + self.residuals['p25'])),
            p50=round(p50),
            p75=round(math.exp(center + self.residuals['p75'])),
            confidence=confidence_for(p50),
            factors=self._factors(plate, factor_limit),
            model_version=self.version,
        )

    def _factors(self, plate: Plate, limit: int) -> list[Factor]:
        """Что тянет цену вверх, по убыванию вклада."""
        found = []
        for key in plate_keys(plate, self.weights):
            weight = self.weights.get(key, 0.0)
            if weight > 0:
                found.append(Factor(_label(key), round(math.exp(weight), 2)))
        if plate.digits_eq_region and self.numeric.get('dg_eq_reg', 0.0) > 0:
            found.append(
                Factor('цифры совпадают с регионом', round(math.exp(self.numeric['dg_eq_reg']), 2))
            )
        found.sort(key=lambda item: item.weight, reverse=True)
        return found[:limit]

    # --- хранение ---

    def to_dict(self) -> dict:
        return {
            'version': self.version,
            'intercept': self.intercept,
            'reference': self.reference,
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
            metrics=payload.get('metrics', {}),
            config=payload.get('config', {}),
            vocabulary=payload.get('vocabulary', {}),
        )


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


def estimate_to_dict(estimate: Estimate) -> dict:
    payload = asdict(estimate)
    payload['price'] = {
        'p25': payload.pop('p25'),
        'p50': payload.pop('p50'),
        'p75': payload.pop('p75'),
    }
    return payload

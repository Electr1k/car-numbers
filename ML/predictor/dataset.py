"""Загрузка выгрузки предложений и подготовка обучающей выборки"""

from dataclasses import dataclass, field

import numpy as np
import pandas as pd

from .plate import CANONICAL_PATTERN, digit_class, letter_class, normalize

# Границы отсечения мусора: 1 ₽ и 2,1 млрд встречаются в базе как ошибки продавцов
MIN_PRICE = 5_000
MAX_PRICE = 20_000_000

# Постоянная времени затухания веса; полураспад получается ~253 дня
DECAY_DAYS = 365.0

# Сбор gosnomeru поверх цены продавца: фиксированная часть и процент
GOSNOMERU_FEE = 20_000
GOSNOMERU_RATE = 0.04


@dataclass(frozen=True)
class Config:
    """Гиперпараметры выборки и модели."""

    min_price: int = MIN_PRICE
    max_price: int = MAX_PRICE
    decay_days: float = DECAY_DAYS
    gosnomeru_fee: int = GOSNOMERU_FEE
    gosnomeru_rate: float = GOSNOMERU_RATE
    # Пороги эффективного (взвешенного) числа наблюдений для собственного коэффициента
    min_series_weight: float = 25.0
    min_region_weight: float = 40.0
    min_series_region_weight: float = 15.0
    alpha: float = 10.0
    # Ворота считаем на горизонте переобучения, интервалы — на длинном: он консервативнее
    holdout_days: int = 30
    interval_days: int = 90

    def as_dict(self) -> dict:
        return {
            'min_price': self.min_price,
            'max_price': self.max_price,
            'decay_days': self.decay_days,
            'gosnomeru_fee': self.gosnomeru_fee,
            'gosnomeru_rate': self.gosnomeru_rate,
            'min_series_weight': self.min_series_weight,
            'min_region_weight': self.min_region_weight,
            'min_series_region_weight': self.min_series_region_weight,
            'alpha': self.alpha,
            'holdout_days': self.holdout_days,
            'interval_days': self.interval_days,
        }


@dataclass
class Dataset:
    """Готовая к обучению выборка и статистика отсева."""

    frame: pd.DataFrame
    as_of: pd.Timestamp
    dropped: dict[str, int] = field(default_factory=dict)

    def __len__(self) -> int:
        return len(self.frame)


def load(path, config: Config | None = None) -> Dataset:
    """Читает CSV выгрузки и готовит колонки признаков, веса и отчёт об отсеве."""
    config = config or Config()
    raw = pd.read_csv(
        path,
        dtype={'number': str, 'type': str, 'provider': str, 'status': str},
    )
    total = len(raw)

    df = raw[raw.type == 'car'].copy()
    dropped_type = total - len(df)

    df['number'] = df.number.map(normalize)
    canonical = df.number.str.fullmatch(CANONICAL_PATTERN)
    dropped_format = int((~canonical).sum())
    df = df[canonical]

    df = df.copy()
    df['price'] = df.price.astype(float)

    # Диапазон проверяем дважды, чтобы отличить мусорную цену от отсева самой дефляцией
    priced = df.price.between(config.min_price, config.max_price)
    deflated = deflate(df, config)
    in_range = df.price.between(config.min_price, config.max_price)

    # Дефляция цену только снижает, поэтому она может вернуть в диапазон то, что было выше потолка
    dropped_price = int((~priced & ~in_range).sum())
    dropped_by_fee = int((priced & ~in_range).sum())
    df = df[in_range].copy()

    df['posted_at'] = pd.to_datetime(df.posted_at, utc=True, format='%Y-%m-%dT%H:%M:%SZ')
    as_of = df.posted_at.max()

    df['digits'] = df.number.str[1:4]
    df['region'] = df.number.str[6:]
    df['series'] = df.number.str[0] + df.number.str[4:6]

    # Классы считаем по уникальным значениям: их сотни, а строк сотни тысяч
    df['digit_class'] = df.digits.map({d: str(digit_class(d)) for d in df.digits.unique()})
    df['letter_class'] = df.series.map({s: str(letter_class(s)) for s in df.series.unique()})

    df['digits_value'] = df.digits.astype(int)
    df['digits_eq_region'] = (df.digits_value == df.region.astype(int)).astype(float)
    df['log_price'] = np.log(df.price)
    df['weight'] = decay_weights(df.posted_at, as_of, config.decay_days)

    return Dataset(
        frame=df,
        as_of=as_of,
        dropped={
            'total_rows': total,
            'not_car': dropped_type,
            'bad_format': dropped_format,
            'price_out_of_range': dropped_price,
            'deflated': deflated,
            'dropped_by_fee': dropped_by_fee,
            'kept': len(df),
        },
    )


def deflate(frame: pd.DataFrame, config: Config) -> int:
    """Снимает сбор gosnomeru, приводя цены всех площадок к цене продавца"""
    rows = frame.provider == 'gosnomeru'
    frame.loc[rows, 'price'] = (
        (frame.loc[rows, 'price'] - config.gosnomeru_fee) / (1 + config.gosnomeru_rate)
    )
    return int(rows.sum())


def decay_weights(posted_at: pd.Series, as_of: pd.Timestamp, decay_days: float) -> np.ndarray:
    """Вес наблюдения по давности публикации."""
    age_days = (as_of - posted_at).dt.total_seconds() / 86_400.0
    return np.exp(-age_days.clip(lower=0) / decay_days).to_numpy()


def effective_size(weights: np.ndarray) -> float:
    """Эффективный размер выборки: сумма весов."""
    return float(np.sum(weights))

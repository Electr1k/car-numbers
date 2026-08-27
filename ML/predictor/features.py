"""Признаки модели: имена, словарь уровней и сборка матрицы"""

from dataclasses import dataclass

import numpy as np
import pandas as pd
from scipy import sparse

from .plate import Plate, digit_class, letter_class

# Числовые признаки: их вес умножается на значение, а не на индикатор
NUMERIC = ('dg_eq_reg', 'log_dnum')

RARE_REGION = 'reg=RARE'


@dataclass(frozen=True)
class Vocabulary:
    """Уровни, заслужившие собственный коэффициент."""

    series: frozenset[str]
    regions: frozenset[str]
    series_regions: frozenset[str]

    def sizes(self) -> dict[str, int]:
        return {
            'series': len(self.series),
            'regions': len(self.regions),
            'series_regions': len(self.series_regions),
        }


def build_vocabulary(frame: pd.DataFrame, config) -> Vocabulary:
    """Отбор уровней по эффективному числу наблюдений, а не по сырому."""
    weight = frame.weight

    def passing(keys: pd.Series, threshold: float) -> frozenset[str]:
        totals = weight.groupby(keys).sum()
        return frozenset(totals[totals >= threshold].index)

    series_regions = frame.series + '|' + frame.region
    return Vocabulary(
        series=passing(frame.series, config.min_series_weight),
        regions=passing(frame.region, config.min_region_weight),
        series_regions=passing(series_regions, config.min_series_region_weight),
    )


def frame_keys(frame: pd.DataFrame, vocab: Vocabulary) -> list[pd.Series]:
    """Категориальные имена признаков для каждой строки выборки."""
    series_key = np.where(
        frame.series.isin(vocab.series),
        'series=' + frame.series,
        'series=RARE_' + frame.letter_class,
    )
    region_key = np.where(
        frame.region.isin(vocab.regions),
        'reg=' + frame.region,
        RARE_REGION,
    )
    combo = frame.series + '|' + frame.region
    cross_key = np.where(combo.isin(vocab.series_regions), 'sxr=' + combo, '')

    return [
        pd.Series(region_key, index=frame.index),
        pd.Series('dg=' + frame.digits, index=frame.index),
        pd.Series(series_key, index=frame.index),
        pd.Series('dcls=' + frame.digit_class, index=frame.index),
        pd.Series('scls=' + frame.letter_class, index=frame.index),
        pd.Series('prov=' + frame.provider, index=frame.index),
        pd.Series(cross_key, index=frame.index),
    ]


def encode(frame: pd.DataFrame, vocab: Vocabulary, names: list[str] | None = None):
    """Разреженная матрица признаков. Возвращает матрицу и порядок имён."""
    keys = frame_keys(frame, vocab)

    if names is None:
        seen: set[str] = set()
        for column in keys:
            seen.update(column.unique())
        seen.discard('')
        names = sorted(seen) + list(NUMERIC)

    index = {name: position for position, name in enumerate(names)}
    rows, cols, values = [], [], []
    row_numbers = np.arange(len(frame))

    for column in keys:
        mapped = column.map(index).to_numpy()
        present = ~pd.isna(mapped)
        rows.append(row_numbers[present])
        cols.append(mapped[present].astype(np.int64))
        values.append(np.ones(int(present.sum())))

    numeric = {
        'dg_eq_reg': frame.digits_eq_region.to_numpy(dtype=float),
        'log_dnum': np.log1p(frame.digits_value.to_numpy(dtype=float)),
    }
    for name, column in numeric.items():
        rows.append(row_numbers)
        cols.append(np.full(len(frame), index[name], dtype=np.int64))
        values.append(column)

    matrix = sparse.csr_matrix(
        (np.concatenate(values), (np.concatenate(rows), np.concatenate(cols))),
        shape=(len(frame), len(names)),
    )
    return matrix, names


def plate_keys(plate: Plate, weights: dict[str, float], provider: str | None = None) -> list[str]:
    """Категориальные признаки одного номера с откатом на класс для редких уровней."""
    series_key = f'series={plate.series}'
    if series_key not in weights:
        series_key = f'series=RARE_{letter_class(plate.series)}'

    region_key = f'reg={plate.region}'
    if region_key not in weights:
        region_key = RARE_REGION

    keys = [
        region_key,
        f'dg={plate.digits}',
        series_key,
        f'dcls={digit_class(plate.digits)}',
        f'scls={letter_class(plate.series)}',
    ]

    cross_key = f'sxr={plate.series}|{plate.region}'
    if cross_key in weights:
        keys.append(cross_key)

    if provider is not None:
        keys.append(f'prov={provider}')

    return keys


def plate_numeric(plate: Plate) -> dict[str, float]:
    """Числовые признаки одного номера."""
    return {
        'dg_eq_reg': float(plate.digits_eq_region),
        'log_dnum': float(np.log1p(plate.digits_value)),
    }

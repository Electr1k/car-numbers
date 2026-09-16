"""Обучение модели цены ГРЗ из выгрузки

    python fit.py --snapshot data/offers-YYYY-mm-dd.csv
"""

import argparse
import sys
from pathlib import Path

import numpy as np
import pandas as pd
from sklearn.linear_model import Ridge

from predictor import dataset, features, model as artifact

# Ворота активации: систематический сдвиг и допустимая просадка точности
SHIFT_LIMITS = (0.9, 1.1)
MAX_MDAPE_REGRESSION = 0.03


def fit_weights(frame: pd.DataFrame, vocab: features.Vocabulary, alpha: float):
    """Гребневая регрессия по логарифму цены с весами по давности."""
    matrix, names = features.encode(frame, vocab)
    ridge = Ridge(alpha=alpha, solver='sparse_cg')
    ridge.fit(matrix, frame.log_price, sample_weight=frame.weight)
    return ridge, matrix, names


def reference_level(frame: pd.DataFrame, weights: dict) -> float:
    """Средневзвешенный вклад площадки — то, что подставляется вместо неё при предсказании."""
    shares = frame.groupby('provider').weight.sum() / frame.weight.sum()
    return sum(share * weights.get(f'prov={name}', 0.0) for name, share in shares.items())


def predict(ridge, matrix, holdout: pd.DataFrame, weights: dict, reference: float) -> np.ndarray:
    """Предсказание в боевом режиме: площадка заменена на эталонный уровень."""
    effect = holdout.provider.map(lambda name: weights.get(f'prov={name}', 0.0)).to_numpy()
    return ridge.predict(matrix) - effect + reference


def score(actual: pd.Series, predicted: np.ndarray) -> dict:
    """Метрики качества на отложенной выборке."""
    predicted_price = np.exp(predicted)
    errors = np.abs(predicted_price - actual.to_numpy()) / actual.to_numpy()
    residuals = np.log(actual.to_numpy()) - predicted
    return {
        'rows': int(len(actual)),
        'mdape': float(np.median(errors)),
        'within_30pct': float(np.mean(errors <= 0.30)),
        'shift': float(np.exp(np.median(residuals))),
    }


def validate(data: dataset.Dataset, config: dataset.Config, days: int) -> tuple[dict, dict]:
    """Обучение на данных до T−days и проверка на последних днях."""
    frame = data.frame
    cutoff = data.as_of - pd.Timedelta(days=days)
    train = frame[frame.posted_at < cutoff]
    holdout = frame[frame.posted_at >= cutoff]
    if holdout.empty:
        raise SystemExit('в выгрузке нет свежих объявлений для проверки')

    vocab = features.build_vocabulary(train, config)
    ridge, _, names = fit_weights(train, vocab, config.alpha)
    weights = dict(zip(names, ridge.coef_))
    matrix, _ = features.encode(holdout, vocab, names)
    predicted = predict(ridge, matrix, holdout, weights, reference_level(train, weights))

    metrics = score(holdout.price, predicted)
    metrics['days'] = days
    metrics['cutoff'] = cutoff.date().isoformat()
    metrics['train_rows'] = int(len(train))

    residuals = np.log(holdout.price.to_numpy()) - predicted
    quantiles = {
        f'p{int(q * 100)}': float(np.quantile(residuals, q))
        for q in (0.10, 0.25, 0.50, 0.75, 0.90)
    }
    return metrics, quantiles


def build(data: dataset.Dataset, config: dataset.Config, version: str) -> artifact.Model:
    """Проверка качества, затем финальный прогон на всей выборке."""
    metrics, _ = validate(data, config, config.holdout_days)
    long_metrics, quantiles = validate(data, config, config.interval_days)
    metrics['long_horizon'] = long_metrics

    frame = data.frame
    vocab = features.build_vocabulary(frame, config)
    ridge, _, names = fit_weights(frame, vocab, config.alpha)
    weights = dict(zip(names, (round(float(w), 5) for w in ridge.coef_)))

    numeric = {name: weights.pop(name) for name in features.NUMERIC}

    # Площадка — свойство витрины, а не номера: при предсказании берём средний уровень
    reference = reference_level(frame, weights)

    metrics['effective_sample'] = round(dataset.effective_size(frame.weight.to_numpy()), 1)
    metrics['rows_total'] = int(len(frame))

    return artifact.Model(
        version=version,
        intercept=float(ridge.intercept_),
        weights=weights,
        numeric=numeric,
        reference=round(float(reference), 5),
        residuals=quantiles,
        centering=centering(frame, vocab, weights, numeric),
        metrics=metrics,
        config=config.as_dict() | {'as_of': data.as_of.date().isoformat()},
        vocabulary=vocab.sizes() | {'features': len(names)},
    )


def centering(frame: pd.DataFrame, vocab: features.Vocabulary, weights: dict,
              numeric: dict) -> dict:
    """Типичный уровень каждой группы признаков — точка отсчёта для раскладки."""
    weight = frame.weight.to_numpy()
    total = weight.sum()
    levels = {}
    for group, column in zip(features.GROUPS, features.frame_keys(frame, vocab)):
        if group == 'prov':
            continue
        contribution = column.map(lambda key: weights.get(key, 0.0)).to_numpy()
        levels[group] = round(float((contribution * weight).sum() / total), 5)

    magnitude = np.log1p(frame.digits_value.to_numpy())
    levels['log_dnum'] = round(
        float(numeric['log_dnum'] * (magnitude * weight).sum() / total), 5
    )
    return levels


def gate(model: artifact.Model, previous: artifact.Model | None) -> list[str]:
    """Причины не активировать версию. Пустой список — можно активировать."""
    problems = []
    shift = model.metrics['shift']
    if not SHIFT_LIMITS[0] <= shift <= SHIFT_LIMITS[1]:
        problems.append(f'сдвиг {shift:.2f}x вне допустимых {SHIFT_LIMITS}')

    if previous and 'mdape' in previous.metrics:
        regression = model.metrics['mdape'] - previous.metrics['mdape']
        if regression > MAX_MDAPE_REGRESSION:
            problems.append(
                f'MdAPE хуже версии {previous.version} на {regression:.1%}'
                f' (порог {MAX_MDAPE_REGRESSION:.0%})'
            )
    return problems


def report(data: dataset.Dataset, model: artifact.Model) -> None:
    dropped = data.dropped
    print(f'выгрузка:      {dropped["total_rows"]} строк')
    print(
        f'  отсеяно:     не car {dropped["not_car"]},'
        f' формат {dropped["bad_format"]},'
        f' цена вне диапазона {dropped["price_out_of_range"]}'
    )
    print(
        f'  дефляция:    {dropped["deflated"]} строк gosnomeru приведены к цене продавца,'
        f' из них {dropped["dropped_by_fee"]} ушли под нижнюю границу'
    )
    print(f'  в обучении:  {dropped["kept"]}')
    print(f'эффективный размер выборки: {model.metrics["effective_sample"]:.0f}')
    print(f'признаков:     {model.vocabulary["features"]}  {model.vocabulary}')
    print()
    for label, metrics in (('ворота', model.metrics), ('длинный', model.metrics['long_horizon'])):
        print(
            f'{label:8s} последние {metrics["days"]:>2} дн (с {metrics["cutoff"]},'
            f' {metrics["rows"]:>5} строк):'
            f'  MdAPE {metrics["mdape"]:.3f}'
            f'  <=30% {metrics["within_30pct"]:.3f}'
            f'  сдвиг {metrics["shift"]:.2f}x'
        )
    spread = model.residuals
    print(
        f'интервал p25..p75: x{np.exp(spread["p25"]):.2f} .. x{np.exp(spread["p75"]):.2f}'
        f'   p10..p90: x{np.exp(spread["p10"]):.2f} .. x{np.exp(spread["p90"]):.2f}'
    )


def main(argv=None) -> int:
    parser = argparse.ArgumentParser(description='обучение модели цены ГРЗ')
    parser.add_argument('--snapshot', type=Path, required=True, help='CSV выгрузки предложений')
    parser.add_argument('--models-dir', type=Path, default=artifact.Registry.MODELS_DIR)
    parser.add_argument('--version', help='имя версии, по умолчанию дата выгрузки')
    parser.add_argument('--force', action='store_true', help='активировать вопреки воротам')
    parser.add_argument('--no-activate', action='store_true', help='только сохранить версию')
    args = parser.parse_args(argv)

    config = dataset.Config()
    data = dataset.load(args.snapshot, config)
    version = args.version or data.as_of.date().isoformat()

    model = build(data, config, version)
    path = artifact.save(model, args.models_dir)
    report(data, model)
    print(f'\nсохранено: {path} ({path.stat().st_size / 1024:.0f} КБ)')

    if args.no_activate:
        return 0

    previous_version = artifact.active_version(args.models_dir)
    previous = artifact.load_active(args.models_dir) if previous_version else None
    problems = gate(model, previous)

    if problems and not args.force:
        for problem in problems:
            print(f'НЕ АКТИВИРОВАНА: {problem}')
        print(f'активной остаётся {previous_version or "— нет активной версии"}')
        return 1

    if problems:
        print('ворота пройдены принудительно (--force):', '; '.join(problems))
    artifact.activate(version, args.models_dir)
    print(f'активна версия {version}')
    return 0


if __name__ == '__main__':
    sys.exit(main())

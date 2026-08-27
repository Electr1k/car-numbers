import statistics

from fastapi import FastAPI, Query
from fastapi.responses import JSONResponse

from predictor import PlateError, expand, mask_count
from predictor.model import Registry, confidence_for

# Больше двух неуказанных знаков — разброс оценки достигает 14 крат, отвечать нечем
MAX_MASKS = 2

app = FastAPI(title='Оценка цены ГРЗ', version='1')
registry = Registry()


def error(status: int, reason: str, message: str) -> JSONResponse:
    return JSONResponse(status_code=status, content={'error': reason, 'message': message})


@app.get('/health')
def health() -> dict:
    model = registry.current()
    return {
        'status': 'ok' if model else 'no_model',
        'model_version': model.version if model else None,
        'models_dir': str(registry.models_dir),
    }


@app.get('/api/v1/predict')
def predict(number: str = Query(min_length=1, max_length=32, description='ГРЗ, маска — *')):
    model = registry.current()
    if model is None:
        return error(503, 'model_unavailable', 'активная версия модели не загружена')

    masks = mask_count(number)
    if masks > MAX_MASKS:
        return error(
            422,
            'too_many_masks',
            f'в номере {masks} неуказанных знаков, оценка возможна не более чем при {MAX_MASKS}',
        )

    try:
        plates = expand(number)
    except PlateError as err:
        status = 422 if err.reason == PlateError.MOTO else 400
        return error(status, err.reason, str(err))

    estimates = [model.estimate(plate) for plate in plates]

    if masks == 0:
        payload = _single(estimates[0])
    else:
        payload = _masked(number, estimates)

    payload['basis'] = {
        'model_version': model.version,
        'trained_on': model.config.get('as_of'),
    }
    return payload


def _single(estimate) -> dict:
    return {
        'number': estimate.number,
        'price': {'p25': estimate.p25, 'p50': estimate.p50, 'p75': estimate.p75},
        'confidence': estimate.confidence,
        'factors': [{'name': factor.name, 'weight': factor.weight} for factor in estimate.factors],
    }


def _masked(number: str, estimates: list) -> dict:
    """Оценка по всем подстановкам: интервал расширен так, чтобы покрыть разброс."""
    middles = sorted(estimate.p50 for estimate in estimates)
    low = min(estimate.p25 for estimate in estimates)
    high = max(estimate.p75 for estimate in estimates)
    median = round(statistics.median(middles))

    return {
        'number': number.strip().upper(),
        'price': {'p25': low, 'p50': median, 'p75': high},
        'confidence': 'low',
        'factors': [],
        'mask': {
            'variants': len(estimates),
            'cheapest': middles[0],
            'dearest': middles[-1],
            'unmasked_confidence': confidence_for(median),
        },
    }

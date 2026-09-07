# Контракт API фронтенда

Сервис фронтенда (BFF) — отдельный от `data-service`. Он ходит в `data-service` за
объявлениями и в `ml-api` за оценкой, склеивает и отдаёт Nuxt-приложению.
В `data-service` фронтенд не ходит напрямую никогда.

```
Nuxt ──► frontend-api (BFF) ──┬──► data-service   (объявления, номера)
                              └──► ml-api         (оценка)
```

Моки в `mock/` собраны **из боевой базы и настоящей модели** скриптом
`generate-mocks.py` — это не выдуманные данные, а реальные строки со всеми
их пробелами. Пересобрать: `ML/venv/bin/python frontend/api/generate-mocks.py`.

## Общее

- База: `/api/v1`
- Деньги — целые рубли, без копеек.
- Даты — ISO 8601 в UTC: `2026-09-04T00:00:00Z`.
- Пагинация — курсорная: `limit` + `cursor`, в ответе `cursor` (`null` — конец).
- Ошибка: `{"error": {"code": "...", "message": "..."}}`, человекочитаемое сообщение
  готово к показу (см. чек-лист доступности: ошибка объясняет, что делать).

### Перечисления

| Поле | Значения |
|---|---|
| `whereabouts` | `on-car`, `on-storage`, `null` |
| `provider` | `gosnomeru`, `autonomera`, `anomera` |
| `confidence` | `high`, `medium`, `low` |
| `status` | `active`, `archive`, `sold` |

## Объекты

### `NumberCard` — карточка в ленте, выдаче и «похожих»

```json
{
  "id": "01a07d18-ed82-75e2-8fa4-99626d7c1545",
  "number": "Т555АА777",
  "region": {
    "code": "777",
    "name": "Москва"
  },
  "price": 1000000,
  "type": "car",
  "reissue_included": null,
  "offers_count": 1,
  "updated_at": "2026-09-01T00:00:00Z"
}
```

- `price` — **минимальная** цена среди активных объявлений номера.
- `reissue_included` — комплектация **того объявления, чья цена показана**, а не всего номера.
  Это важно: в 27,8% многоофферных карточек минимум идёт без переоформления,
  и подпись обязана это сказать.
- Сортировка по дате обновления — по `refreshed_at`, то есть по минимальной.

### `Offer` — объявление на странице номера

```json
{
  "id": "01a07d2c-11dc-703d-add6-9535c9d01267",
  "provider": "anomera",
  "price": 60000,
  "status": "active",
  "reissue_included": true,
  "whereabouts": null,
  "views": 5,
  "comment": "+ переоформление 30ТЫС( через мое авто) , С переоформлением помогу . Связь 8999999999 Владислав Север МО, без лишних записей в ПТС",
  "posted_at": "2026-07-11T00:00:00Z",
  "refreshed_at": "2026-07-11T00:00:00Z",
  "url": "https://anomera6.ru/gos-nomera-v-moskovskoy-oblasti/1000003248-x515kb550.html"
}
```

`whereabouts`, `views` и `comment` — **nullable**. Правило показа: есть значение —
показываем, `null` — строка не отрисовывается. `views` приходит `null`, когда
у площадки нет счётчика или он равен нулю (у gosnomeru всегда ноль — это не «мало
смотрят», это «мы не знаем»).

### `Estimate` — оценка

```json
{
  "p25": 51268, "p50": 64459, "p75": 81660,
  "confidence": "high",
  "breakdown": {
    "base": 118732,
    "items": [
      { "code": "region", "value": "550", "title": "Регион 550",
        "multiplier": 0.7314, "exact": true }
    ]
  },
  "basis": { "model_version": "2026-09-04", "trained_on": "2026-09-04" }
}
```

Произведение `base` на все `multiplier` даёт `p50` с точностью округления.
`estimate` может быть `null` — когда ml-api недоступен; страница обязана
показать объявления и без оценки.

Для номеров с маской добавляется блок `mask`:
`{ "variants": 10, "cheapest": 737958, "dearest": 3461022 }`, `confidence`
принудительно `low`.

## Эндпоинты

### `GET /api/v1/feed`

Лента главной.

| Параметр | По умолчанию | Что |
|---|---|---|
| `limit` | 12 | 1–48 |
| `cursor` | — | из `cursor` |
| `region` | — | необязателен |

```json
{ "items": [NumberCard], "next_cursor": "ZmVlZDoy" }
```

### `GET /api/v1/search`

| Параметр | Что |
|---|---|
| `q` | номер или маска: `А123ВС777`, `А*2*ВС77` |
| `region` | код региона, необязателен |
| `price_min`, `price_max` | рубли |
| `categories` | `same_digits`, `round_hundred`, `first_ten`, `same_letters`, `mirror`, `ladder` |
| `reissue_included` | `true`, `false`, `null` |
| `sort` | `price_asc`, `price_desc`, `updated_desc` (по умолчанию) |
| `limit`, `cursor` | пагинация |

```json
{ "query": {...}, "total": 34, "items": [NumberCard], "cursor": null }
```

`total` — сколько всего подошло; нужен для заголовка «34 номера по маске».
Пустой `items` — не ошибка, а штатный исход: фронт показывает предложение
поставить триггер.

### `GET /api/v1/plates/{id}`

Одна страница — один запрос. Состояние определяет сервер.

```json
{
  "id": "01a07d2c-11dc-703d-add6-9535c9d01261",
  "number": "Х515КВ550",
  "region": {
    "code": "550",
    "name": "Московская обл."
  },
  "active_offers": [Offer],
  "archive_offers": null,
  "similar": []
}
```
### `GET /api/v1/regions`

Справочник для фильтра.

```json
{ "items": [ { "codes": [797, 77], "name": "Москва" } ] }
```

### `GET /api/v1/valuation?number=X001XX01`

Оценка цены номера авто.

```json
{
  "number": "Х515КВ550",
  "region": {
    "code": "550",
    "name": "Московская обл."
  },
  "p25": 45477,
  "p50": 57213,
  "p75": 72513,
  "confidence": "high",
  "breakdown": {
    "base": 118732,
    "items": [
      {
        "code": "series",
        "value": "ХКВ",
        "title": "Серия ХКВ",
        "multiplier": 0.9607,
        "exact": true
      },
      {
        "code": "region",
        "value": "550",
        "title": "Регион 550",
        "multiplier": 0.8831,
        "exact": true
      },
      {
        "code": "digits",
        "value": "515",
        "title": "Цифры 515",
        "multiplier": 0.864,
        "exact": true
      },
      {
        "code": "letter_class",
        "value": "OTHER",
        "title": "Разные буквы",
        "multiplier": 0.8162,
        "exact": true
      },
      {
        "code": "digit_class",
        "value": "ABA",
        "title": "Зеркальные цифры",
        "multiplier": 0.8054,
        "exact": true
      }
    ]
  },
  "basis": {
    "model_version": "2026-09-04",
    "trained_on": "2026-09-04"
  }
}
```

## Отложено

- `POST /api/v1/triggers` — требует профиля, состав не определён.
- Авторизация и `/api/v1/me`.
- Нейропоиск — пункт меню есть, внутри заглушка.

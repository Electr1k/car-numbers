CREATE TABLE IF NOT EXISTS features (
    id UUID PRIMARY KEY,
    key VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    active BOOL NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT features_key_uk UNIQUE (key)
);

comment on column features.id is 'Идентификатор (UUIDv7)';
comment on column features.key is 'Ключ фичи';
comment on column features.name is 'Название фичи';
comment on column features.active is 'Флаг активности фичи';

INSERT INTO features (id, key, name)
VALUES(
    '01a00b6d-bdad-76db-901f-5e3794b8c96e',
    'import-autonomera-offers',
    'Импорт предложений из autonomera777'
)

INSERT INTO features (id, key, name)
VALUES(
    '01a00b73-9ac2-7148-ba5c-478b058d035e',
    'dispatch-import-offer-detail',
    'Вызов импорта предложений из autonomera777'
)

INSERT INTO features (id, key, name)
VALUES(
    '01a00b6f-a3bf-72b5-9797-1b99d1945d7b',
    'import-offer-detail',
    'Импорт детальной информации об оффере (из любого провайдера)'
)

INSERT INTO features (id, key, name)
VALUES(
    '01a00b72-26c0-709c-be8f-bac499ff5a46',
    'dispatch-import-offer-detail',
    'Вызов импорта детальной информации об оффере (из любого провайдера)'
)

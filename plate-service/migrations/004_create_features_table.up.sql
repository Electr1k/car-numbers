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
) ON CONFLICT (key) DO NOTHING;

INSERT INTO features (id, key, name)
VALUES(
    '01a00b73-9ac2-7148-ba5c-478b058d035e',
    'dispatch-import-autonomera-offers',
    'Вызов импорта предложений из autonomera777'
) ON CONFLICT (key) DO NOTHING;

INSERT INTO features (id, key, name)
VALUES(
    '01a00b6f-a3bf-72b5-9797-1b99d1945d7b',
    'import-offer-detail',
    'Импорт детальной информации об оффере (из любого провайдера)'
) ON CONFLICT (key) DO NOTHING;

INSERT INTO features (id, key, name)
VALUES(
    '01a00b72-26c0-709c-be8f-bac499ff5a46',
    'dispatch-import-offer-detail',
    'Вызов импорта детальной информации об оффере (из любого провайдера)'
) ON CONFLICT (key) DO NOTHING;

INSERT INTO features (id, key, name)
VALUES(
    '01a03b15-05fa-7724-ba04-fc785194556d',
    'import-gosnomeru-offers',
    'Импорт предложений из gosnomeru'
) ON CONFLICT (key) DO NOTHING;

INSERT INTO features (id, key, name)
VALUES(
    '01a0401a-7f44-776e-aa63-4130be56acdd',
    'dispatch-import-gosnomeru-offers',
    'Вызов импорта предложений из gosnomeru'
) ON CONFLICT (key) DO NOTHING;

INSERT INTO features (id, key, name)
VALUES(
    '01a05a47-6f4d-7474-8e55-d4b2504a097c',
    'import-anomera-offers',
    'Импорт предложений из anomera'
) ON CONFLICT (key) DO NOTHING;

INSERT INTO features (id, key, name)
VALUES(
    '01a05a47-a247-711c-b58f-522c5c2345c6',
    'dispatch-import-anomera-offers',
    'Вызов импорта предложений из anomera'
) ON CONFLICT (key) DO NOTHING;

INSERT INTO features (id, key, name)
VALUES(
    '01a06374-7816-735b-ac98-1ca771b03e4f',
    'sync-active-offers',
    'Синхронизация активных офферов (из всех провайдеров)'
) ON CONFLICT (key) DO NOTHING;

INSERT INTO features (id, key, name)
VALUES(
    '01a063a0-9889-714b-b953-f10d26908d09',
    'dispatch-sync-active-offers',
    'Вызов синхронизации активных офферов (из всех провайдеров)'
) ON CONFLICT (key) DO NOTHING;
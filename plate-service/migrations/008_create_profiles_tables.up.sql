CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    name text NOT NULL,
    phone VARCHAR(16),
    email text,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

comment on column users.id is 'Идентификатор (UUIDv7)';
comment on column users.name is 'Имя';
comment on column users.phone is 'Телефон';
comment on column users.email is 'E-mail';

CREATE TABLE IF NOT EXISTS profiles (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    provider VARCHAR(255) NOT NULL,
    external_id text NOT NULL,
    url VARCHAR(255) NOT NULL,
    login text,
    name text NOT NULL,
    phone VARCHAR(16),
    email text,
    badge VARCHAR(32),
    rating INT,
    registered_at TIMESTAMP WITH TIME ZONE,
    last_visit_at TIMESTAMP WITH TIME ZONE,
    raw text NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT profiles_provider_external_id_uk UNIQUE (provider, external_id)
);

CREATE INDEX IF NOT EXISTS idx_profiles_user_id ON profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

comment on column profiles.id is 'Идентификатор (UUIDv7)';
comment on column profiles.user_id is 'Доменный пользователь';
comment on column profiles.provider is 'Провайдер';
comment on column profiles.external_id is 'Идентификатор пользователя у провайдера';
comment on column profiles.url is 'Ссылка на профиль';
comment on column profiles.login is 'Логин';
comment on column profiles.name is 'Имя, если не заполнено - логин';
comment on column profiles.phone is 'Телефон из профиля';
comment on column profiles.email is 'E-mail из профиля';
comment on column profiles.badge is 'Статус продавца у провайдера';
comment on column profiles.rating is 'Рейтинг у провайдера';
comment on column profiles.registered_at is 'Дата регистрации у провайдера';
comment on column profiles.last_visit_at is 'Дата последнего визита у провайдера';
comment on column profiles.raw is 'Сырой профиль поставщика';

ALTER TABLE offers ADD COLUMN IF NOT EXISTS profile_external_id text;

CREATE INDEX IF NOT EXISTS idx_offers_profile_external_id ON offers(profile_external_id);

comment on column offers.profile_external_id is 'Идентификатор пользователя у провайдера, разместившего предложение';

INSERT INTO features (id, key, name)
VALUES(
    '01a0d525-5b59-76b5-8d1f-e91ab5e72ac1',
    'import-profile',
    'Импорт профиля пользователя (из любого провайдера)'
) ON CONFLICT (key) DO NOTHING;

INSERT INTO features (id, key, name)
VALUES(
    '01a0d525-5b65-7073-989d-9844c0cd3e38',
    'dispatch-import-profile',
    'Вызов импорта профиля пользователя (из любого провайдера)'
) ON CONFLICT (key) DO NOTHING;

INSERT INTO features (id, key, name)
VALUES(
    '01a0d57d-0f80-7806-af31-7ed47390044e',
    'import-autonomera-profiles',
    'Импорт профилей из autonomera777'
) ON CONFLICT (key) DO NOTHING;

INSERT INTO features (id, key, name)
VALUES(
    '01a0d57d-0f8a-767e-a5e8-7eaba33e42a2',
    'dispatch-import-autonomera-profiles',
    'Вызов импорта профилей из autonomera777'
) ON CONFLICT (key) DO NOTHING;

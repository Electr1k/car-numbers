CREATE TABLE IF NOT EXISTS profiles (
    id UUID PRIMARY KEY,
    provider VARCHAR(255) NOT NULL,
    external_id text NOT NULL,
    url VARCHAR(255) NOT NULL,
    icon_url VARCHAR(255),
    login text NOT NULL,
    name text,
    phone VARCHAR(16),
    email text,
    registered_at TIMESTAMP WITH TIME ZONE,
    last_visit_at TIMESTAMP WITH TIME ZONE,
    raw text NOT NULL,
    description text,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT profiles_provider_external_id_uk UNIQUE (provider, external_id)
);

comment on column profiles.id is 'Идентификатор профиля';
comment on column profiles.provider is 'Провайдер, в котором найден профиль';
comment on column profiles.external_id is 'Идентификатор у провайдера';
comment on column profiles.url is 'Ссылка на профиль у провайдера';
comment on column profiles.icon_url is 'Ссылка на иконку профиля';
comment on column profiles.login is 'Логин';
comment on column profiles.name is 'Имя';
comment on column profiles.phone is 'Номер';
comment on column profiles.email is 'Email';
comment on column profiles.registered_at is 'Дата регистрации';
comment on column profiles.last_visit_at is 'Дата последнего визита';
comment on column profiles.raw is 'Сырой объект поставщика';
comment on column profiles.description is 'Описание профиля';

CREATE INDEX IF NOT EXISTS idx_profiles_phone ON profiles(phone);

CREATE INDEX IF NOT EXISTS idx_profiles_email ON profiles(lower(email));

CREATE INDEX IF NOT EXISTS idx_profiles_login ON profiles(lower(login));

ALTER TABLE offers ADD COLUMN IF NOT EXISTS profile_id UUID REFERENCES profiles(id);

CREATE INDEX IF NOT EXISTS idx_offers_profile_id ON offers(profile_id);

comment on column offers.profile_id is 'Профиль продавца';

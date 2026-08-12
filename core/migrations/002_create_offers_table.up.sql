CREATE TABLE IF NOT EXISTS offers (
    id UUID PRIMARY KEY,
    number_id UUID NOT NULL REFERENCES numbers(id),
    provider VARCHAR(255) NOT NULL,
    external_id text NOT NULL,
    price DECIMAL(12,2) NOT NULL,
    status VARCHAR(255) NOT NULL,
    posted_at TIMESTAMP WITH TIME ZONE NOT NULL,
    url VARCHAR(255) NOT NULL,
    raw text NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT offers_provider_external_id_uk UNIQUE (provider, external_id)
);


CREATE INDEX offers_number_id ON offers (number_id);

comment on column offers.id is 'Идентификатор (UUIDv7)';
comment on column offers.number_id is 'Идентификатор номера';
comment on column offers.provider is 'Провайдер, в котором найдено предложение';
comment on column offers.external_id is 'Идентификатор у провайдера';
comment on column offers.price is 'Цена';
comment on column offers.status is 'Статус предложения';
comment on column offers.posted_at is 'Дата публикации';
comment on column offers.url is 'Ссылка на предложение';
comment on column offers.raw is 'Сырой объект поставщика';

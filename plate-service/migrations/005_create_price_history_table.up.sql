CREATE TABLE IF NOT EXISTS price_history (
    id UUID PRIMARY KEY,
    offer_id UUID NOT NULL REFERENCES offers(id),
    plate_id UUID NOT NULL REFERENCES plates(id),
    price DECIMAL(12,2),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_price_history_offer_id ON price_history (offer_id, created_at);

CREATE INDEX idx_price_history_plate_id ON price_history (plate_id, created_at);

comment on column price_history.id is 'Идентификатор (UUIDv7)';
comment on column price_history.offer_id is 'Идентификатор оффера';
comment on column price_history.plate_id is 'Идентификатор номера';
comment on column price_history.price is 'Цена';

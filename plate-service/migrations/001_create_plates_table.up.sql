CREATE TABLE IF NOT EXISTS plates (
    id UUID PRIMARY KEY,
    number VARCHAR(9) NOT NULL,
    type VARCHAR(16) NOT NULL,
    region_code VARCHAR(3) GENERATED ALWAYS AS (
        CASE WHEN substring(number from 7) ~ '^[0-9]{2,3}$' THEN substring(number from 7) END
    ) STORED,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT plates_number_type_uk UNIQUE (number, type)
);

CREATE INDEX IF NOT EXISTS idx_plates_number ON plates(number);

CREATE INDEX IF NOT EXISTS idx_plates_region_code ON plates(region_code);

comment on column plates.id is 'Идентификатор (UUIDv7)';
comment on column plates.number is 'Номер';
comment on column plates.type is 'Тип ТС';
comment on column plates.region_code is 'Код региона из номера';

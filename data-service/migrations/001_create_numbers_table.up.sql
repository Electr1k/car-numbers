CREATE TABLE IF NOT EXISTS numbers (
    id UUID PRIMARY KEY,
    number VARCHAR(9) NOT NULL,
    type VARCHAR(16) NOT NULL,
    region_code VARCHAR(3) GENERATED ALWAYS AS (
        CASE WHEN substring(number from 7) ~ '^[0-9]{2,3}$' THEN substring(number from 7) END
    ) STORED,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT numbers_number_type_uk UNIQUE (number, type)
);

CREATE INDEX IF NOT EXISTS idx_number ON numbers(number);

CREATE INDEX IF NOT EXISTS idx_numbers_region_code ON numbers(region_code);

comment on column numbers.id is 'Идентификатор (UUIDv7)';
comment on column numbers.number is 'Номер';
comment on column numbers.type is 'Тип ТС';
comment on column numbers.region_code is 'Код региона из номера';

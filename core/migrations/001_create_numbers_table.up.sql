CREATE TABLE IF NOT EXISTS numbers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    number VARCHAR(9) NOT NULL,
    type VARCHAR(16) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT numbers_number_type_uk UNIQUE (number, type)
);

CREATE INDEX IF NOT EXISTS idx_number ON numbers(number);

comment on column numbers.id is 'Идентификатор';
comment on column numbers.number is 'Номер';
comment on column numbers.type is 'Тип ТС';

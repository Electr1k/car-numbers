CREATE TABLE IF NOT EXISTS numbers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    number VARCHAR(9) NOT NULL,
    letters VARCHAR(3) NOT NULL,
    digits VARCHAR(3) NOT NULL,
    region VARCHAR(3) NOT NULL,
    type VARCHAR(16) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT numbers_number_uk UNIQUE (number)
);

CREATE INDEX IF NOT EXISTS idx_number ON numbers(number);

comment on column numbers.id is 'Идентификатор';
comment on column numbers.number is 'Номер';
comment on column numbers.letters is 'Буквы номера';
comment on column numbers.digits is 'Цифры номера';
comment on column numbers.region is 'Регион';
comment on column numbers.type is 'Тип (авто/мото/прицеп)';

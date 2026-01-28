CREATE TABLE IF NOT EXISTS car_numbers (
    number VARCHAR(9) PRIMARY KEY,
    price DECIMAL(12,2) NOT NULL,
    posted_at DATE DEFAULT CURRENT_DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_price ON car_numbers(price);
CREATE INDEX IF NOT EXISTS idx_posted_at ON car_numbers(posted_at);

comment on column car_numbers.number is 'Номер авто';
comment on column car_numbers.price is 'Цена';
comment on column car_numbers.posted_at is 'Дата публикации';
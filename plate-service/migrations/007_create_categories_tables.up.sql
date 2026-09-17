CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR(128) PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT categories_name_uk UNIQUE (name)
);

comment on column categories.id is 'Идентификатор категории';
comment on column categories.name is 'Название категории';

CREATE TABLE IF NOT EXISTS plate_categories (
    plate_id UUID NOT NULL REFERENCES plates(id),
    category_id VARCHAR(128) NOT NULL REFERENCES categories(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (plate_id, category_id)
);
CREATE INDEX IF NOT EXISTS idx_plate_categories_category_id ON plate_categories(category_id);

comment on column plate_categories.plate_id is 'Идентификатор номера';
comment on column plate_categories.category_id is 'Код категории';


INSERT INTO categories (id, name)
VALUES
    ('same-letters', 'Одинаковые буквы'),
    ('mirrored-letters', 'Зеркальные буквы'),
    ('pair-letters', 'Пара букв'),
    ('letter-staircase', 'Лесенка букв'),
    ('same-digits', 'Одинаковые цифры'),
    ('first-ten', 'Первая десятка'),
    ('round-hundreds', 'Круглые сотни'),
    ('zero-edges', 'Ноль по краям'),
    ('mirrored-digits', 'Зеркальные цифры'),
    ('pair-digits', 'Пара цифр'),
    ('digits-staircase', 'Лесенка цифр'),
    ('digits-as-region', 'Цифры как регион')
ON CONFLICT (id) DO NOTHING;

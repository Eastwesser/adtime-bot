-- +goose Up
-- Добавляем категории продуктов
INSERT INTO product_categories (name) VALUES 
('Типография'),
('Наклейки'),
('Кожаные изделия');

-- Добавляем продукты для типографии
INSERT INTO products (category_id, name, description) VALUES
(1, 'Визитки', 'Печать визиток на бумаге или картоне'),
(1, 'Бирки', 'Изготовление бирок для одежды'),
(1, 'Листовки', 'Печать листовок А4, А5'),
(1, 'Буклеты', 'Складывающиеся буклеты'),
(1, 'Каталоги', 'Многостраничные каталоги'),
(1, 'Календари', 'Настенные и карманные календари'),
(1, 'Открытки', 'Печать открыток');

-- Добавляем поле product_id в заказы
ALTER TABLE orders ADD COLUMN product_id INTEGER REFERENCES products(id);
ALTER TABLE orders ADD COLUMN quantity INTEGER;
ALTER TABLE orders ADD COLUMN options JSONB;

-- +goose Down
ALTER TABLE orders DROP COLUMN product_id;
ALTER TABLE orders DROP COLUMN quantity;
ALTER TABLE orders DROP COLUMN options;
DELETE FROM products WHERE category_id IN (SELECT id FROM product_categories WHERE name IN ('Типография', 'Наклейки'));
DELETE FROM product_categories WHERE name IN ('Типография', 'Наклейки');
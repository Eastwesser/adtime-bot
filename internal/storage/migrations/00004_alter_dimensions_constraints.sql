-- +goose Up
-- First drop the existing constraints
ALTER TABLE orders 
DROP CONSTRAINT IF EXISTS orders_height_cm_check,
DROP CONSTRAINT IF EXISTS orders_width_cm_check;

-- Add new constraints that allow zero values
ALTER TABLE orders
ADD CONSTRAINT orders_height_cm_check CHECK (height_cm >= 0 AND height_cm <= 50),
ADD CONSTRAINT orders_width_cm_check CHECK (width_cm >= 0 AND width_cm <= 80);

-- +goose Down
-- Restore original constraints
ALTER TABLE orders 
DROP CONSTRAINT IF EXISTS orders_height_cm_check,
DROP CONSTRAINT IF EXISTS orders_width_cm_check;

ALTER TABLE orders
ADD CONSTRAINT orders_height_cm_check CHECK (height_cm > 0 AND height_cm <= 50),
ADD CONSTRAINT orders_width_cm_check CHECK (width_cm > 0 AND width_cm <= 80);
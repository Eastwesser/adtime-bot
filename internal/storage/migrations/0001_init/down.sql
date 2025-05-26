DROP INDEX IF EXISTS idx_textures_in_stock;
DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS textures;
DROP EXTENSION IF EXISTS "uuid-ossp";
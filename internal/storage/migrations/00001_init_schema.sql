-- +goose Up
-- Enable UUID extension if not exists
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


-- Create textures table with improved constraints
CREATE TABLE textures (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(255)   NOT NULL,
    price_per_dm2 DECIMAL(10, 2) NOT NULL CHECK (price_per_dm2 > 0),
    image_url     VARCHAR(512),
    in_stock      BOOLEAN        NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_texture_name UNIQUE (name)
);

-- Create orders table with improved constraints and relations
CREATE SEQUENCE orders_id_seq START 1;
CREATE TABLE orders (
    id         INTEGER PRIMARY KEY DEFAULT nextval('orders_id_seq'),
    user_id    BIGINT         NOT NULL,
    width_cm   INTEGER        NOT NULL CHECK (width_cm > 0 AND width_cm <= 80),
    height_cm  INTEGER        NOT NULL CHECK (height_cm > 0 AND height_cm <= 50),
    texture_id UUID           NOT NULL,
    price      DECIMAL(10, 2) NOT NULL CHECK (price > 0),
    contact    VARCHAR(50)    NOT NULL CHECK (contact ~ '^\+[0-9]{10,15}$'),
    created_at TIMESTAMP      NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP      NOT NULL DEFAULT NOW(),
    status     VARCHAR(20)    NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'processing', 'completed', 'cancelled')),
    
    CONSTRAINT fk_texture
      FOREIGN KEY(texture_id) 
      REFERENCES textures(id)
      ON DELETE RESTRICT
);

-- Create indexes with explicit naming convention
CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_status ON orders (status);
CREATE INDEX idx_orders_created_at ON orders (created_at);
CREATE INDEX idx_textures_in_stock ON textures (in_stock) WHERE in_stock = TRUE;

-- +goose Down
DROP INDEX IF EXISTS idx_textures_in_stock;
DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS textures;
DROP EXTENSION IF EXISTS "uuid-ossp";
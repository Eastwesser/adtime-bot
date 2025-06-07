-- +goose Up
CREATE TABLE users (
    user_id BIGINT PRIMARY KEY,
    agreed_to_tpa BOOLEAN NOT NULL DEFAULT FALSE,
    phone_number VARCHAR(20),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS users;
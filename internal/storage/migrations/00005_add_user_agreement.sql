-- +goose Up
ALTER TABLE users ADD COLUMN agreed_to_tpa BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN phone_number VARCHAR(20);
-- +goose Down
ALTER TABLE users DROP COLUMN agreed_to_tpa;
ALTER TABLE users DROP COLUMN phone_number;
-- +goose Up
ALTER TABLE users ADD COLUMN currency TEXT NOT NULL DEFAULT 'USD';

-- +goose Down
ALTER TABLE users DROP COLUMN currency;

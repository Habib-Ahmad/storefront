-- +goose Up
ALTER TABLE orders ADD COLUMN note TEXT;

-- +goose Down
ALTER TABLE orders DROP COLUMN IF EXISTS note;

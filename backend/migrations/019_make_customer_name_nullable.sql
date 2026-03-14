-- +goose Up
ALTER TABLE orders ALTER COLUMN customer_name DROP NOT NULL;

-- +goose Down
ALTER TABLE orders ALTER COLUMN customer_name SET NOT NULL;

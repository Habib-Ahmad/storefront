-- +goose Up
ALTER TABLE order_items ADD COLUMN cost_price_at_sale DECIMAL(15, 2);

-- +goose Down
ALTER TABLE order_items DROP COLUMN IF EXISTS cost_price_at_sale;

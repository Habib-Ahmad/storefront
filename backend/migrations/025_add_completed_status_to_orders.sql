-- +goose Up
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_fulfillment_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_fulfillment_status_check
    CHECK (fulfillment_status IN ('processing', 'completed', 'shipped', 'delivered', 'cancelled'));

-- +goose Down
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_fulfillment_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_fulfillment_status_check
    CHECK (fulfillment_status IN ('processing', 'shipped', 'delivered', 'cancelled'));
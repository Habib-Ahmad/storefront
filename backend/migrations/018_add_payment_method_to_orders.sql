-- +goose Up
ALTER TABLE orders
    ADD COLUMN payment_method VARCHAR(20) NOT NULL DEFAULT 'online'
        CHECK (payment_method IN ('online', 'cash', 'transfer'));

CREATE INDEX idx_orders_payment_method ON orders(tenant_id, payment_method);

-- +goose Down
DROP INDEX IF EXISTS idx_orders_payment_method;
ALTER TABLE orders DROP COLUMN IF EXISTS payment_method;

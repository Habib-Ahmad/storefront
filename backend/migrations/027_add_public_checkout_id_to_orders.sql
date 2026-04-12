-- +goose Up
ALTER TABLE orders
    ADD COLUMN public_checkout_id UUID;

CREATE UNIQUE INDEX idx_orders_public_checkout_id_active
    ON orders(tenant_id, public_checkout_id)
    WHERE public_checkout_id IS NOT NULL AND payment_status <> 'failed';

-- +goose Down
DROP INDEX IF EXISTS idx_orders_public_checkout_id_active;

ALTER TABLE orders
    DROP COLUMN IF EXISTS public_checkout_id;
-- +goose Up
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_payment_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_payment_status_check
    CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded'));

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_fulfillment_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_fulfillment_status_check
    CHECK (fulfillment_status IN ('processing', 'shipped', 'delivered', 'cancelled'));

ALTER TABLE shipments DROP CONSTRAINT IF EXISTS shipments_status_check;
ALTER TABLE shipments ADD CONSTRAINT shipments_status_check
    CHECK (status IN ('queued', 'picked_up', 'in_transit', 'delivered', 'failed'));

-- +goose Down
ALTER TABLE shipments DROP CONSTRAINT IF EXISTS shipments_status_check;
ALTER TABLE shipments ADD CONSTRAINT shipments_status_check
    CHECK (status IN ('queued', 'picked_up', 'in_transit', 'delivered'));

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_fulfillment_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_fulfillment_status_check
    CHECK (fulfillment_status IN ('processing', 'shipped', 'delivered'));

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_payment_status_check;
ALTER TABLE orders ADD CONSTRAINT orders_payment_status_check
    CHECK (payment_status IN ('pending', 'paid', 'failed'));

-- +goose Up
CREATE TABLE orders (
    id                 UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id          UUID           NOT NULL REFERENCES tenants(id),
    tracking_slug      VARCHAR(255)   NOT NULL UNIQUE,
    is_delivery        BOOLEAN        NOT NULL DEFAULT false,
    customer_name      VARCHAR(255)   NOT NULL,
    customer_phone     VARCHAR(50),
    customer_email     VARCHAR(255),
    shipping_address   TEXT,
    total_amount       DECIMAL(15, 2) NOT NULL,
    shipping_fee       DECIMAL(15, 2) NOT NULL DEFAULT 0,
    payment_status     VARCHAR(50)    NOT NULL DEFAULT 'pending'
                           CHECK (payment_status IN ('pending', 'paid', 'failed')),
    fulfillment_status VARCHAR(50)    NOT NULL DEFAULT 'processing'
                           CHECK (fulfillment_status IN ('processing', 'shipped', 'delivered')),
    created_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_tenant_id          ON orders(tenant_id);
CREATE INDEX idx_orders_tracking_slug      ON orders(tracking_slug);
CREATE INDEX idx_orders_payment_status     ON orders(tenant_id, payment_status);
CREATE INDEX idx_orders_fulfillment_status ON orders(tenant_id, fulfillment_status);

-- +goose Down
DROP INDEX IF EXISTS idx_orders_fulfillment_status;
DROP INDEX IF EXISTS idx_orders_payment_status;
DROP INDEX IF EXISTS idx_orders_tracking_slug;
DROP INDEX IF EXISTS idx_orders_tenant_id;
DROP TABLE IF EXISTS orders;

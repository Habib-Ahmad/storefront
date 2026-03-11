-- +goose Up
CREATE TABLE shipments (
    id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID         NOT NULL UNIQUE REFERENCES orders(id),
    tenant_id       UUID         NOT NULL REFERENCES tenants(id),
    status          VARCHAR(50)  NOT NULL DEFAULT 'queued'
                        CHECK (status IN ('queued', 'picked_up', 'in_transit', 'delivered')),
    carrier_ref     VARCHAR(255),
    tracking_number VARCHAR(255),
    carrier_history JSONB        NOT NULL DEFAULT '[]',
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shipments_tenant_id ON shipments(tenant_id);
CREATE INDEX idx_shipments_status    ON shipments(tenant_id, status);

-- +goose Down
DROP INDEX IF EXISTS idx_shipments_status;
DROP INDEX IF EXISTS idx_shipments_tenant_id;
DROP TABLE IF EXISTS shipments;

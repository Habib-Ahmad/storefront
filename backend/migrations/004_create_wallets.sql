-- +goose Up
CREATE TABLE wallets (
    id                      UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id               UUID           NOT NULL UNIQUE REFERENCES tenants(id),
    available_balance       DECIMAL(15, 2) NOT NULL DEFAULT 0,
    pending_balance         DECIMAL(15, 2) NOT NULL DEFAULT 0,
    last_transaction_id     UUID,
    last_reconciliation_at  TIMESTAMPTZ
);

CREATE INDEX idx_wallets_tenant_id ON wallets(tenant_id);

-- +goose Down
DROP INDEX IF EXISTS idx_wallets_tenant_id;
DROP TABLE IF EXISTS wallets;

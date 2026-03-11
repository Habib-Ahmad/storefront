-- +goose Up
CREATE TABLE products (
    id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID         NOT NULL REFERENCES tenants(id),
    name         VARCHAR(255) NOT NULL,
    description  TEXT,
    category     VARCHAR(100),
    is_available BOOLEAN      NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_products_tenant_id ON products(tenant_id)            WHERE deleted_at IS NULL;
CREATE INDEX idx_products_category  ON products(tenant_id, category)  WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_products_category;
DROP INDEX IF EXISTS idx_products_tenant_id;
DROP TABLE IF EXISTS products;

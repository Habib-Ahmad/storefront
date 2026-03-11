-- +goose Up
CREATE TABLE product_variants (
    id         UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID           NOT NULL REFERENCES products(id),
    sku        VARCHAR(255)   NOT NULL UNIQUE,
    attributes JSONB          NOT NULL DEFAULT '{}',
    price      DECIMAL(15, 2) NOT NULL,
    stock_qty  INTEGER,
    is_default BOOLEAN        NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_product_variants_product_id ON product_variants(product_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_product_variants_sku        ON product_variants(sku);

-- +goose Down
DROP INDEX IF EXISTS idx_product_variants_sku;
DROP INDEX IF EXISTS idx_product_variants_product_id;
DROP TABLE IF EXISTS product_variants;

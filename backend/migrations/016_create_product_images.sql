-- +goose Up
ALTER TABLE products DROP COLUMN IF EXISTS image_url;

CREATE TABLE product_images (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID        NOT NULL REFERENCES products(id),
    url        TEXT        NOT NULL,
    sort_order INTEGER     NOT NULL DEFAULT 0,
    is_primary BOOLEAN     NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_images_product_id ON product_images(product_id);

-- +goose Down
DROP INDEX IF EXISTS idx_product_images_product_id;
DROP TABLE IF EXISTS product_images;

ALTER TABLE products ADD COLUMN image_url TEXT;

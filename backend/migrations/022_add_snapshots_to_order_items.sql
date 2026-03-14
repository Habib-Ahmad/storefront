-- +goose Up
ALTER TABLE order_items
    ADD COLUMN product_name  VARCHAR(255),
    ADD COLUMN variant_label VARCHAR(255);

-- +goose Down
ALTER TABLE order_items
    DROP COLUMN IF EXISTS variant_label,
    DROP COLUMN IF EXISTS product_name;

-- +goose Up
CREATE TABLE order_items (
    id            UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id      UUID           NOT NULL REFERENCES orders(id),
    variant_id    UUID           NOT NULL REFERENCES product_variants(id),
    quantity      INTEGER        NOT NULL CHECK (quantity > 0),
    price_at_sale DECIMAL(15, 2) NOT NULL
);

CREATE INDEX idx_order_items_order_id   ON order_items(order_id);
CREATE INDEX idx_order_items_variant_id ON order_items(variant_id);

-- +goose Down
DROP INDEX IF EXISTS idx_order_items_variant_id;
DROP INDEX IF EXISTS idx_order_items_order_id;
DROP TABLE IF EXISTS order_items;

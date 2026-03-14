-- +goose Up
ALTER TABLE product_variants ADD COLUMN cost_price DECIMAL(15, 2);

-- +goose Down
ALTER TABLE product_variants DROP COLUMN IF EXISTS cost_price;

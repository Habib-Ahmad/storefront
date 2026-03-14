-- +goose Up
ALTER TABLE product_variants DROP CONSTRAINT product_variants_sku_key;
ALTER TABLE product_variants ADD CONSTRAINT product_variants_product_id_sku_key UNIQUE (product_id, sku);

-- +goose Down
ALTER TABLE product_variants DROP CONSTRAINT product_variants_product_id_sku_key;
ALTER TABLE product_variants ADD CONSTRAINT product_variants_sku_key UNIQUE (sku);
a
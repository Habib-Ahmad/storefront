-- +goose Up
ALTER TABLE product_images
    ADD CONSTRAINT uq_product_images_sort_order UNIQUE (product_id, sort_order);

-- +goose Down
ALTER TABLE product_images
    DROP CONSTRAINT IF EXISTS uq_product_images_sort_order;

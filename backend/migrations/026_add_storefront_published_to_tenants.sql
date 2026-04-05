-- +goose Up
ALTER TABLE tenants
    ADD COLUMN storefront_published BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE tenants
    DROP COLUMN IF EXISTS storefront_published;
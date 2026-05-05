-- +goose Up
ALTER TABLE tenants DROP COLUMN IF EXISTS active_modules;

-- +goose Down
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS active_modules JSONB NOT NULL DEFAULT '{}';
-- +goose Up

-- User profile fields
ALTER TABLE users
    ADD COLUMN first_name VARCHAR(100),
    ADD COLUMN last_name  VARCHAR(100),
    ADD COLUMN phone      VARCHAR(50);

-- Tenant profile fields
ALTER TABLE tenants
    ADD COLUMN contact_email VARCHAR(255),
    ADD COLUMN contact_phone VARCHAR(50),
    ADD COLUMN address       TEXT,
    ADD COLUMN logo_url      TEXT;

-- +goose Down
ALTER TABLE tenants
    DROP COLUMN IF EXISTS logo_url,
    DROP COLUMN IF EXISTS address,
    DROP COLUMN IF EXISTS contact_phone,
    DROP COLUMN IF EXISTS contact_email;

ALTER TABLE users
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS last_name,
    DROP COLUMN IF EXISTS first_name;

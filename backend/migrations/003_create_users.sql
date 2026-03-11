-- +goose Up
CREATE TABLE users (
    id          UUID         PRIMARY KEY,
    tenant_id   UUID         NOT NULL REFERENCES tenants(id),
    email       VARCHAR(255) NOT NULL,
    role        VARCHAR(50)  NOT NULL DEFAULT 'staff'
                    CHECK (role IN ('admin', 'staff', 'manager')),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_users_tenant_id ON users(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_email     ON users(email)     WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_tenant_id;
DROP TABLE IF EXISTS users;

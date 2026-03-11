-- +goose Up
CREATE TABLE tenants (
    id                      UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    tier_id                 UUID         NOT NULL REFERENCES tiers(id),
    name                    VARCHAR(255) NOT NULL,
    slug                    VARCHAR(255) NOT NULL UNIQUE,
    paystack_subaccount_id  VARCHAR(255),
    active_modules          JSONB        NOT NULL DEFAULT '{}',
    status                  VARCHAR(50)  NOT NULL DEFAULT 'active'
                                CHECK (status IN ('active', 'suspended')),
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ
);

CREATE INDEX idx_tenants_slug   ON tenants(slug)   WHERE deleted_at IS NULL;
CREATE INDEX idx_tenants_status ON tenants(status) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_tenants_status;
DROP INDEX IF EXISTS idx_tenants_slug;
DROP TABLE IF EXISTS tenants;

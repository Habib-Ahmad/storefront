-- +goose Up
CREATE TABLE tiers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(100)   NOT NULL,
    debt_ceiling    DECIMAL(15, 2) NOT NULL DEFAULT 0,
    commission_rate DECIMAL(5, 4)  NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

-- Seed default tiers
INSERT INTO tiers (name, debt_ceiling, commission_rate)
VALUES
    ('Standard', 0,      0.05),
    ('Premium',  50000,  0.03);

-- +goose Down
DROP TABLE IF EXISTS tiers;

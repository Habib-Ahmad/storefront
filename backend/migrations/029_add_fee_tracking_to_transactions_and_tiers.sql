-- +goose Up
ALTER TABLE tiers
    ADD COLUMN IF NOT EXISTS commission_cap DECIMAL(15, 2) NOT NULL DEFAULT 2000;

UPDATE tiers
SET commission_cap = 2000
WHERE commission_cap <= 0;

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS platform_fee_base DECIMAL(15, 2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS platform_fee_rate DECIMAL(8, 6) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS platform_fee_cap DECIMAL(15, 2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS platform_fee_amount DECIMAL(15, 2) NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_transactions_platform_fee_amount
    ON transactions (platform_fee_amount)
    WHERE platform_fee_amount > 0;

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_platform_fee_amount;

ALTER TABLE transactions
    DROP COLUMN IF EXISTS platform_fee_amount,
    DROP COLUMN IF EXISTS platform_fee_cap,
    DROP COLUMN IF EXISTS platform_fee_rate,
    DROP COLUMN IF EXISTS platform_fee_base;

ALTER TABLE tiers
    DROP COLUMN IF EXISTS commission_cap;
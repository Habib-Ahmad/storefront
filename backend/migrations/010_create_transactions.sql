-- +goose Up

-- last_transaction_id on wallets deliberately has no FK here to avoid a circular reference;
-- the application layer keeps that pointer up to date.
CREATE TABLE transactions (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id       UUID           NOT NULL REFERENCES wallets(id),
    order_id        UUID           REFERENCES orders(id),
    amount          DECIMAL(15, 2) NOT NULL,
    running_balance DECIMAL(15, 2) NOT NULL,
    type            VARCHAR(50)    NOT NULL
                        CHECK (type IN ('credit', 'debit', 'commission', 'payout')),
    signature       TEXT           NOT NULL,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE RULE transactions_no_update AS ON UPDATE TO transactions DO INSTEAD NOTHING;
CREATE RULE transactions_no_delete AS ON DELETE TO transactions DO INSTEAD NOTHING;

CREATE INDEX idx_transactions_wallet_id   ON transactions(wallet_id);
CREATE INDEX idx_transactions_order_id    ON transactions(order_id) WHERE order_id IS NOT NULL;
CREATE INDEX idx_transactions_created_at  ON transactions(wallet_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_created_at;
DROP INDEX IF EXISTS idx_transactions_order_id;
DROP INDEX IF EXISTS idx_transactions_wallet_id;
DROP RULE IF EXISTS transactions_no_delete ON transactions;
DROP RULE IF EXISTS transactions_no_update ON transactions;
DROP TABLE IF EXISTS transactions;

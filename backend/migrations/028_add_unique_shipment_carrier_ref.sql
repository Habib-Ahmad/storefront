-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS idx_shipments_carrier_ref_unique
    ON shipments(carrier_ref)
    WHERE carrier_ref IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_shipments_carrier_ref_unique;
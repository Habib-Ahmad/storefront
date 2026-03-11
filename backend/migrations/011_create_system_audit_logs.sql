-- +goose Up

-- Partitioned by month. New partitions must be added manually each month (or via pg_partman).
CREATE TABLE system_audit_logs (
    id          BIGSERIAL    NOT NULL,
    tenant_id   UUID         NOT NULL REFERENCES tenants(id),
    user_id     UUID,
    action      VARCHAR(100) NOT NULL,
    diff        JSONB        NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE system_audit_logs_2026_03
    PARTITION OF system_audit_logs
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');

CREATE TABLE system_audit_logs_2026_04
    PARTITION OF system_audit_logs
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

CREATE TABLE system_audit_logs_2026_05
    PARTITION OF system_audit_logs
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');

CREATE INDEX idx_audit_logs_tenant_id  ON system_audit_logs(tenant_id, created_at DESC);
CREATE INDEX idx_audit_logs_action     ON system_audit_logs(action, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_tenant_id;
DROP TABLE IF EXISTS system_audit_logs;

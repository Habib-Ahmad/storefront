-- +goose Up
CREATE OR REPLACE FUNCTION create_audit_log_partition(target_date DATE)
RETURNS VOID LANGUAGE plpgsql AS $$
DECLARE
    partition_name  TEXT;
    partition_start DATE;
    partition_end   DATE;
BEGIN
    partition_start := DATE_TRUNC('month', target_date)::DATE;
    partition_end   := (partition_start + INTERVAL '1 month')::DATE;
    partition_name  := 'system_audit_logs_' || TO_CHAR(partition_start, 'YYYY_MM');

    IF NOT EXISTS (
        SELECT 1 FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE c.relname = partition_name AND n.nspname = 'public'
    ) THEN
        EXECUTE FORMAT(
            'CREATE TABLE %I PARTITION OF system_audit_logs FOR VALUES FROM (%L) TO (%L)',
            partition_name, partition_start, partition_end
        );
    END IF;
END;
$$;

-- +goose Down
DROP FUNCTION IF EXISTS create_audit_log_partition(DATE);

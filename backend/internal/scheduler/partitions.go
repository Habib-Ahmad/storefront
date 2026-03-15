package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EnsureAuditLogPartitions creates partitions for the current and next month.
// It calls the DB function created by migration 012.
func EnsureAuditLogPartitions(ctx context.Context, pool *pgxpool.Pool) error {
	now := time.Now()
	for _, t := range []time.Time{now, now.AddDate(0, 1, 0)} {
		if _, err := pool.Exec(ctx, "SELECT create_audit_log_partition($1::date)", t); err != nil {
			return fmt.Errorf("audit log partition %s: %w", t.Format("2006-01"), err)
		}
	}
	return nil
}

// RunMonthlyPartitioner runs EnsureAuditLogPartitions on the 1st of each month.
// It blocks until ctx is cancelled, allowing clean shutdown.
func RunMonthlyPartitioner(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			if t.Day() == 1 {
				if err := EnsureAuditLogPartitions(ctx, pool); err != nil {
					log.Error("partition scheduler", "error", err)
				}
			}
		}
	}
}

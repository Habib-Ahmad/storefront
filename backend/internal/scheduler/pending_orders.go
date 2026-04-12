package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"storefront/backend/internal/service"
)

const pendingOrderSweepInterval = 5 * time.Minute
const pendingOrderSweepBatchSize = 100

// RunPendingOrderExpiry expires stale unpaid online orders on startup and on a short interval.
func RunPendingOrderExpiry(ctx context.Context, pool *pgxpool.Pool, paymentSvc *service.PaymentService, ttl time.Duration, log *slog.Logger) {
	if ttl <= 0 {
		return
	}

	if err := sweepPendingOrders(ctx, pool, paymentSvc, ttl, log); err != nil {
		log.Error("pending-order expiry (startup)", "error", err)
	}

	ticker := time.NewTicker(pendingOrderSweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := sweepPendingOrders(ctx, pool, paymentSvc, ttl, log); err != nil {
				log.Error("pending-order expiry", "error", err)
			}
		}
	}
}

func sweepPendingOrders(ctx context.Context, pool *pgxpool.Pool, paymentSvc *service.PaymentService, ttl time.Duration, log *slog.Logger) error {
	for {
		expired, err := paymentSvc.SweepExpiredPendingOrders(ctx, pool, ttl, pendingOrderSweepBatchSize)
		if err != nil {
			return err
		}
		if expired == 0 {
			return nil
		}
		log.Info("expired stale pending orders", "count", expired, "ttl", ttl.String())
	}
}

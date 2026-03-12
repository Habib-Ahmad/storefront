package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"storefront/backend/internal/service"
)

// RunDailyChainVerifier runs VerifyChain for every active tenant's wallet once a day.
// On chain tamper, VerifyChain suspends the tenant and writes an audit log entry.
func RunDailyChainVerifier(ctx context.Context, pool *pgxpool.Pool, walletSvc *service.WalletService) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := verifyAllChains(ctx, pool, walletSvc); err != nil {
				slog.Error("chain verifier", "error", err)
			}
		}
	}
}

func verifyAllChains(ctx context.Context, pool *pgxpool.Pool, walletSvc *service.WalletService) error {
	rows, err := pool.Query(ctx, `
		SELECT w.id, w.tenant_id
		FROM wallets w
		JOIN tenants t ON t.id = w.tenant_id
		WHERE t.status = 'active' AND t.deleted_at IS NULL`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var walletID, tenantID uuid.UUID
		if err := rows.Scan(&walletID, &tenantID); err != nil {
			slog.Error("chain verifier scan", "error", err)
			continue
		}
		if err := walletSvc.VerifyChain(ctx, walletID, tenantID); err != nil {
			slog.Warn("chain verification failed", "tenant_id", tenantID, "wallet_id", walletID, "error", err)
		}
	}
	return rows.Err()
}

package scheduler

import (
	"context"
	"log/slog"
	"time"

	"storefront/backend/internal/repository"
	"storefront/backend/internal/service"
)

// RunDailyChainVerifier runs VerifyChain for every active tenant's wallet once a day.
// On chain tamper, VerifyChain suspends the tenant and writes an audit log entry.
func RunDailyChainVerifier(ctx context.Context, wallets repository.WalletRepository, walletSvc *service.WalletService, log *slog.Logger) {
	// Run once immediately on startup.
	if err := verifyAllChains(ctx, wallets, walletSvc, log); err != nil {
		log.Error("chain verifier (startup)", "error", err)
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := verifyAllChains(ctx, wallets, walletSvc, log); err != nil {
				log.Error("chain verifier", "error", err)
			}
		}
	}
}

func verifyAllChains(ctx context.Context, wallets repository.WalletRepository, walletSvc *service.WalletService, log *slog.Logger) error {
	activeWallets, err := wallets.ListActiveWallets(ctx)
	if err != nil {
		return err
	}

	for _, aw := range activeWallets {
		if err := walletSvc.VerifyChain(ctx, aw.WalletID, aw.TenantID); err != nil {
			log.Warn("chain verification failed", "tenant_id", aw.TenantID, "wallet_id", aw.WalletID, "error", err)
		}
	}
	return nil
}

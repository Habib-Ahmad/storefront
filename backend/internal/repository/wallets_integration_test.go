package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

func TestWalletRepositoryCreateGetAndUpdateBalances(t *testing.T) {
	ctx := context.Background()
	pool := setupRepositoryTestDB(t)
	tenantID := createTenantFixture(t, ctx, pool, models.TenantStatusActive)

	repo := repository.NewWalletRepository(pool)
	wallet := &models.Wallet{TenantID: tenantID}
	if err := repo.Create(ctx, wallet); err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	if wallet.ID == uuid.Nil {
		t.Fatal("expected wallet id to be set")
	}
	if !wallet.AvailableBalance.Equal(decimal.Zero) {
		t.Fatalf("available_balance: want 0, got %s", wallet.AvailableBalance)
	}
	if !wallet.PendingBalance.Equal(decimal.Zero) {
		t.Fatalf("pending_balance: want 0, got %s", wallet.PendingBalance)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	lockedWallet, err := repo.WithTx(tx).GetByTenantIDForUpdate(ctx, tenantID)
	if err != nil {
		tx.Rollback(ctx)
		t.Fatalf("get wallet for update: %v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback tx: %v", err)
	}
	if lockedWallet.ID != wallet.ID {
		t.Fatalf("locked wallet id: want %s, got %s", wallet.ID, lockedWallet.ID)
	}

	lastTransactionID := uuid.New()
	wallet.AvailableBalance = decimal.NewFromInt(1900)
	wallet.PendingBalance = decimal.NewFromInt(250)
	wallet.LastTransactionID = &lastTransactionID
	if err := repo.UpdateBalances(ctx, wallet); err != nil {
		t.Fatalf("update wallet balances: %v", err)
	}

	persisted, err := repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		t.Fatalf("get wallet: %v", err)
	}
	if !persisted.AvailableBalance.Equal(wallet.AvailableBalance) {
		t.Fatalf("available_balance: want %s, got %s", wallet.AvailableBalance, persisted.AvailableBalance)
	}
	if !persisted.PendingBalance.Equal(wallet.PendingBalance) {
		t.Fatalf("pending_balance: want %s, got %s", wallet.PendingBalance, persisted.PendingBalance)
	}
	if persisted.LastTransactionID == nil || *persisted.LastTransactionID != lastTransactionID {
		t.Fatalf("last_transaction_id: want %s, got %v", lastTransactionID, persisted.LastTransactionID)
	}
}

func TestWalletRepositoryListActiveWalletsFiltersSuspendedTenants(t *testing.T) {
	ctx := context.Background()
	pool := setupRepositoryTestDB(t)
	activeTenantID := createTenantFixture(t, ctx, pool, models.TenantStatusActive)
	suspendedTenantID := createTenantFixture(t, ctx, pool, models.TenantStatusSuspended)

	repo := repository.NewWalletRepository(pool)
	activeWallet := &models.Wallet{TenantID: activeTenantID}
	if err := repo.Create(ctx, activeWallet); err != nil {
		t.Fatalf("create active wallet: %v", err)
	}
	suspendedWallet := &models.Wallet{TenantID: suspendedTenantID}
	if err := repo.Create(ctx, suspendedWallet); err != nil {
		t.Fatalf("create suspended wallet: %v", err)
	}

	activeWallets, err := repo.ListActiveWallets(ctx)
	if err != nil {
		t.Fatalf("list active wallets: %v", err)
	}
	if len(activeWallets) != 1 {
		t.Fatalf("expected 1 active wallet, got %d", len(activeWallets))
	}
	if activeWallets[0].WalletID != activeWallet.ID {
		t.Fatalf("wallet_id: want %s, got %s", activeWallet.ID, activeWallets[0].WalletID)
	}
	if activeWallets[0].TenantID != activeTenantID {
		t.Fatalf("tenant_id: want %s, got %s", activeTenantID, activeWallets[0].TenantID)
	}
}

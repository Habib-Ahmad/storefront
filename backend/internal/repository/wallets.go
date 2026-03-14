package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"storefront/backend/internal/db"
	"storefront/backend/internal/models"
)

type WalletRepository interface {
	Create(ctx context.Context, w *models.Wallet) error
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*models.Wallet, error)
	GetByTenantIDForUpdate(ctx context.Context, tenantID uuid.UUID) (*models.Wallet, error)
	UpdateBalances(ctx context.Context, w *models.Wallet) error
	WithTx(tx db.DBTX) WalletRepository
}

type walletRepo struct {
	db   db.DBTX
	pool *pgxpool.Pool
}

func NewWalletRepository(pool *pgxpool.Pool) WalletRepository {
	return &walletRepo{db: pool, pool: pool}
}

func (r *walletRepo) WithTx(tx db.DBTX) WalletRepository {
	return &walletRepo{db: tx, pool: r.pool}
}

func (r *walletRepo) Create(ctx context.Context, w *models.Wallet) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO wallets (tenant_id) VALUES ($1)
		RETURNING id, available_balance, pending_balance`,
		w.TenantID,
	).Scan(&w.ID, &w.AvailableBalance, &w.PendingBalance)
}

func (r *walletRepo) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*models.Wallet, error) {
	w := &models.Wallet{}
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, available_balance, pending_balance,
		       last_transaction_id, last_reconciliation_at
		FROM wallets WHERE tenant_id = $1`, tenantID,
	).Scan(&w.ID, &w.TenantID, &w.AvailableBalance, &w.PendingBalance,
		&w.LastTransactionID, &w.LastReconciliationAt)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *walletRepo) GetByTenantIDForUpdate(ctx context.Context, tenantID uuid.UUID) (*models.Wallet, error) {
	w := &models.Wallet{}
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, available_balance, pending_balance,
		       last_transaction_id, last_reconciliation_at
		FROM wallets WHERE tenant_id = $1 FOR UPDATE`, tenantID,
	).Scan(&w.ID, &w.TenantID, &w.AvailableBalance, &w.PendingBalance,
		&w.LastTransactionID, &w.LastReconciliationAt)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (r *walletRepo) UpdateBalances(ctx context.Context, w *models.Wallet) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE wallets
		SET available_balance    = $1,
		    pending_balance      = $2,
		    last_transaction_id  = $3
		WHERE id = $4 AND tenant_id = $5`,
		w.AvailableBalance, w.PendingBalance, w.LastTransactionID, w.ID, w.TenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("wallet %s not found", w.ID)
	}
	return nil
}

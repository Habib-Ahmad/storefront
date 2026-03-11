package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"storefront/backend/internal/models"
)

type WalletRepository interface {
	Create(ctx context.Context, w *models.Wallet) error
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*models.Wallet, error)
	UpdateBalances(ctx context.Context, w *models.Wallet) error
}

type walletRepo struct{ db *pgxpool.Pool }

func NewWalletRepository(db *pgxpool.Pool) WalletRepository {
	return &walletRepo{db: db}
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

func (r *walletRepo) UpdateBalances(ctx context.Context, w *models.Wallet) error {
	_, err := r.db.Exec(ctx, `
		UPDATE wallets
		SET available_balance    = $1,
		    pending_balance      = $2,
		    last_transaction_id  = $3
		WHERE id = $4`,
		w.AvailableBalance, w.PendingBalance, w.LastTransactionID, w.ID)
	return err
}

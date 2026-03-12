package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"storefront/backend/internal/models"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *models.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	ListByWallet(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]models.Transaction, error)
	ListByWalletAsc(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]models.Transaction, error)
	GetLatestByWallet(ctx context.Context, walletID uuid.UUID) (*models.Transaction, error)
}

type transactionRepo struct{ db *pgxpool.Pool }

func NewTransactionRepository(db *pgxpool.Pool) TransactionRepository {
	return &transactionRepo{db: db}
}

func (r *transactionRepo) Create(ctx context.Context, tx *models.Transaction) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO transactions (wallet_id, order_id, amount, running_balance, type, signature)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`,
		tx.WalletID, tx.OrderID, tx.Amount, tx.RunningBalance, tx.Type, tx.Signature,
	).Scan(&tx.ID, &tx.CreatedAt)
}

func (r *transactionRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	tx := &models.Transaction{}
	err := r.db.QueryRow(ctx, `
		SELECT id, wallet_id, order_id, amount, running_balance, type, signature, created_at
		FROM transactions WHERE id = $1`, id,
	).Scan(&tx.ID, &tx.WalletID, &tx.OrderID, &tx.Amount, &tx.RunningBalance,
		&tx.Type, &tx.Signature, &tx.CreatedAt)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (r *transactionRepo) ListByWallet(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]models.Transaction, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, wallet_id, order_id, amount, running_balance, type, signature, created_at
		FROM transactions
		WHERE wallet_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		walletID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(&tx.ID, &tx.WalletID, &tx.OrderID, &tx.Amount,
			&tx.RunningBalance, &tx.Type, &tx.Signature, &tx.CreatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, rows.Err()
}

func (r *transactionRepo) ListByWalletAsc(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]models.Transaction, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, wallet_id, order_id, amount, running_balance, type, signature, created_at
		FROM transactions
		WHERE wallet_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3`,
		walletID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(&tx.ID, &tx.WalletID, &tx.OrderID, &tx.Amount,
			&tx.RunningBalance, &tx.Type, &tx.Signature, &tx.CreatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, rows.Err()
}

func (r *transactionRepo) GetLatestByWallet(ctx context.Context, walletID uuid.UUID) (*models.Transaction, error) {
	tx := &models.Transaction{}
	err := r.db.QueryRow(ctx, `
		SELECT id, wallet_id, order_id, amount, running_balance, type, signature, created_at
		FROM transactions
		WHERE wallet_id = $1
		ORDER BY created_at DESC LIMIT 1`, walletID,
	).Scan(&tx.ID, &tx.WalletID, &tx.OrderID, &tx.Amount, &tx.RunningBalance,
		&tx.Type, &tx.Signature, &tx.CreatedAt)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

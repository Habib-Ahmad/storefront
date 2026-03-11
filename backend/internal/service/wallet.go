package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var (
	ErrChainTampered  = errors.New("ledger chain integrity violation")
	ErrWalletNotFound = errors.New("wallet not found")
)

type WalletService struct {
	wallets      repository.WalletRepository
	transactions repository.TransactionRepository
	tenants      repository.TenantRepository
	secret       string
}

func NewWalletService(
	wallets repository.WalletRepository,
	transactions repository.TransactionRepository,
	tenants repository.TenantRepository,
	secret string,
) *WalletService {
	return &WalletService{wallets: wallets, transactions: transactions, tenants: tenants, secret: secret}
}

// Credit adds amount to the wallet's pending balance and appends a signed ledger entry.
func (s *WalletService) Credit(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal, orderID *uuid.UUID) (*models.Transaction, error) {
	return s.record(ctx, walletID, amount, models.TransactionTypeCredit, orderID)
}

// Debit subtracts amount from available balance and appends a signed ledger entry.
func (s *WalletService) Debit(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal, orderID *uuid.UUID) (*models.Transaction, error) {
	return s.record(ctx, walletID, amount.Neg(), models.TransactionTypeDebit, orderID)
}

// ReleasePending moves amount from pending to available (post-delivery settlement).
func (s *WalletService) ReleasePending(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal) error {
	w, err := s.wallets.GetByTenantID(ctx, walletID)
	if err != nil {
		return ErrWalletNotFound
	}
	w.PendingBalance = w.PendingBalance.Sub(amount)
	w.AvailableBalance = w.AvailableBalance.Add(amount)
	return s.wallets.UpdateBalances(ctx, w)
}

// VerifyChain re-computes every HMAC for the wallet and returns ErrChainTampered
// if any entry does not match. On mismatch it also suspends the tenant.
func (s *WalletService) VerifyChain(ctx context.Context, walletID uuid.UUID, tenantID uuid.UUID) error {
	const pageSize = 500
	offset := 0
	var prevSig string

	for {
		txs, err := s.transactions.ListByWallet(ctx, walletID, pageSize, offset)
		if err != nil {
			return fmt.Errorf("list transactions: %w", err)
		}
		if len(txs) == 0 {
			break
		}

		// Transactions are returned newest-first; reverse for chain verification.
		for i := len(txs) - 1; i >= 0; i-- {
			tx := txs[i]
			expected := computeSignature(tx.Amount, tx.RunningBalance, prevSig, s.secret)
			if !hmac.Equal([]byte(tx.Signature), []byte(expected)) {
				_ = s.suspendTenant(ctx, tenantID)
				return ErrChainTampered
			}
			prevSig = tx.Signature
		}

		if len(txs) < pageSize {
			break
		}
		offset += pageSize
	}
	return nil
}

// record is the single write path for all ledger entries.
func (s *WalletService) record(ctx context.Context, walletID uuid.UUID, amount decimal.Decimal, txType models.TransactionType, orderID *uuid.UUID) (*models.Transaction, error) {
	w, err := s.wallets.GetByTenantID(ctx, walletID)
	if err != nil {
		return nil, ErrWalletNotFound
	}

	var prevSig string
	if w.LastTransactionID != nil {
		prev, err := s.transactions.GetLatestByWallet(ctx, w.ID)
		if err != nil {
			return nil, fmt.Errorf("fetch chain tip: %w", err)
		}
		prevSig = prev.Signature
	}

	newBalance := w.AvailableBalance.Add(amount)
	sig := computeSignature(amount, newBalance, prevSig, s.secret)

	tx := &models.Transaction{
		WalletID:       w.ID,
		OrderID:        orderID,
		Amount:         amount,
		RunningBalance: newBalance,
		Type:           txType,
		Signature:      sig,
	}
	if err := s.transactions.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	w.AvailableBalance = newBalance
	w.LastTransactionID = &tx.ID
	if err := s.wallets.UpdateBalances(ctx, w); err != nil {
		return nil, fmt.Errorf("update wallet balances: %w", err)
	}

	return tx, nil
}

func (s *WalletService) suspendTenant(ctx context.Context, tenantID uuid.UUID) error {
	t, err := s.tenants.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}
	t.Status = models.TenantStatusSuspended
	return s.tenants.Update(ctx, t)
}

// computeSignature returns HMAC-SHA256(amount|running_balance|prev_sig, secret).
func computeSignature(amount, runningBalance decimal.Decimal, prevSig, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(amount.String()))
	mac.Write([]byte(runningBalance.String()))
	mac.Write([]byte(prevSig))
	return hex.EncodeToString(mac.Sum(nil))
}

package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/apperr"
	"storefront/backend/internal/db"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var (
	ErrChainTampered       = apperr.New(http.StatusInternalServerError, "ledger chain integrity violation")
	ErrWalletNotFound      = apperr.NotFound("wallet not found")
	ErrDebtCeilingExceeded = apperr.Unprocessable("debt ceiling exceeded")
)

type WalletService struct {
	wallets      repository.WalletRepository
	transactions repository.TransactionRepository
	tenants      repository.TenantRepository
	tiers        repository.TierRepository
	auditLogs    repository.AuditLogRepository
	pool         TxBeginner
	secret       string
}

// TxBeginner starts a database transaction. Satisfied by *pgxpool.Pool.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

func NewWalletService(
	wallets repository.WalletRepository,
	transactions repository.TransactionRepository,
	tenants repository.TenantRepository,
	secret string,
) *WalletService {
	return &WalletService{wallets: wallets, transactions: transactions, tenants: tenants, secret: secret}
}

func (s *WalletService) SetPool(pool TxBeginner) { s.pool = pool }

// SetTierRepo injects the tier repository after construction (avoids circular init order).
func (s *WalletService) SetTierRepo(tiers repository.TierRepository) { s.tiers = tiers }

// SetAuditLogRepo injects the audit-log repository after construction.
func (s *WalletService) SetAuditLogRepo(al repository.AuditLogRepository) { s.auditLogs = al }

// Credit adds funds to the wallet's pending balance and appends a signed ledger entry.
// Funds stay in pending until delivery is confirmed; use ReleasePending to make them available.
func (s *WalletService) Credit(ctx context.Context, tenantID uuid.UUID, amount decimal.Decimal, orderID *uuid.UUID) (*models.Transaction, error) {
	return s.record(ctx, tenantID, amount, models.TransactionTypeCredit, true, orderID)
}

// CreditWithTx performs a Credit inside an externally-managed database transaction.
// The caller MUST commit/rollback the tx.
func (s *WalletService) CreditWithTx(ctx context.Context, tx db.DBTX, tenantID uuid.UUID, amount decimal.Decimal, orderID *uuid.UUID) (*models.Transaction, error) {
	wallets := s.wallets.WithTx(tx)
	transactions := s.transactions.WithTx(tx)
	return s.doRecordForUpdate(ctx, wallets, transactions, tenantID, amount, models.TransactionTypeCredit, true, orderID)
}

// CreditAvailableWithTx performs a CreditAvailable inside an externally-managed database transaction.
// The caller MUST commit/rollback the tx.
func (s *WalletService) CreditAvailableWithTx(ctx context.Context, tx db.DBTX, tenantID uuid.UUID, amount decimal.Decimal, orderID *uuid.UUID) (*models.Transaction, error) {
	wallets := s.wallets.WithTx(tx)
	transactions := s.transactions.WithTx(tx)
	return s.doRecordForUpdate(ctx, wallets, transactions, tenantID, amount, models.TransactionTypeCredit, false, orderID)
}

// RecordCommissionWithTx appends a commission deduction entry inside an externally-managed transaction.
func (s *WalletService) RecordCommissionWithTx(ctx context.Context, tx db.DBTX, tenantID uuid.UUID, amount decimal.Decimal, orderID *uuid.UUID) (*models.Transaction, error) {
	wallets := s.wallets.WithTx(tx)
	transactions := s.transactions.WithTx(tx)
	return s.doRecordForUpdate(ctx, wallets, transactions, tenantID, amount.Neg(), models.TransactionTypeCommission, false, orderID)
}

// CreditAvailable adds funds directly to available balance (for offline/cash/transfer sales).
func (s *WalletService) CreditAvailable(ctx context.Context, tenantID uuid.UUID, amount decimal.Decimal, orderID *uuid.UUID) (*models.Transaction, error) {
	return s.record(ctx, tenantID, amount, models.TransactionTypeCredit, false, orderID)
}

// Debit subtracts amount from available balance and appends a signed ledger entry.
// It enforces the tier debt ceiling: available_balance - amount must not fall below -debtCeiling.
func (s *WalletService) Debit(ctx context.Context, tenantID uuid.UUID, amount decimal.Decimal, orderID *uuid.UUID) (*models.Transaction, error) {
	w, err := s.wallets.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, ErrWalletNotFound
	}
	if s.tiers == nil {
		return nil, fmt.Errorf("tier repository not configured")
	}
	tenant, err := s.tenants.GetByID(ctx, w.TenantID)
	if err != nil {
		return nil, fmt.Errorf("get tenant for ceiling check: %w", err)
	}
	tier, err := s.tiers.GetByID(ctx, tenant.TierID)
	if err != nil {
		return nil, fmt.Errorf("get tier for ceiling check: %w", err)
	}
	newBalance := w.AvailableBalance.Sub(amount)
	if newBalance.LessThan(tier.DebtCeiling.Neg()) {
		return nil, ErrDebtCeilingExceeded
	}
	return s.record(ctx, tenantID, amount.Neg(), models.TransactionTypeDebit, false, orderID)
}

// RecordCommission appends a commission deduction entry to the ledger (available side, not escrow).
func (s *WalletService) RecordCommission(ctx context.Context, tenantID uuid.UUID, amount decimal.Decimal, orderID *uuid.UUID) (*models.Transaction, error) {
	return s.record(ctx, tenantID, amount.Neg(), models.TransactionTypeCommission, false, orderID)
}

// Refund debits the wallet for a cancelled order's refund.
func (s *WalletService) Refund(ctx context.Context, tenantID uuid.UUID, amount decimal.Decimal, orderID *uuid.UUID) (*models.Transaction, error) {
	return s.record(ctx, tenantID, amount.Neg(), models.TransactionTypeRefund, false, orderID)
}

// ReleasePending moves amount from pending to available (post-delivery settlement).
// Records a zero-amount ledger entry (release doesn't change net flow, only reclassifies).
func (s *WalletService) ReleasePending(ctx context.Context, tenantID uuid.UUID, amount decimal.Decimal, orderID *uuid.UUID) error {
	doRelease := func(wallets repository.WalletRepository, transactions repository.TransactionRepository) error {
		var w *models.Wallet
		var err error
		if s.pool != nil {
			w, err = wallets.GetByTenantIDForUpdate(ctx, tenantID)
		} else {
			w, err = wallets.GetByTenantID(ctx, tenantID)
		}
		if err != nil {
			return ErrWalletNotFound
		}

		var prevSig string
		var prevRunningBalance decimal.Decimal
		if w.LastTransactionID != nil {
			prev, err := transactions.GetLatestByWallet(ctx, w.ID)
			if err != nil {
				return fmt.Errorf("fetch chain tip: %w", err)
			}
			prevSig = prev.Signature
			prevRunningBalance = prev.RunningBalance
		}

		sig := computeSignature(decimal.Zero, prevRunningBalance, prevSig, s.secret)
		tx := &models.Transaction{
			WalletID:       w.ID,
			OrderID:        orderID,
			Amount:         decimal.Zero,
			RunningBalance: prevRunningBalance,
			Type:           models.TransactionTypeRelease,
			Signature:      sig,
		}
		if err := transactions.Create(ctx, tx); err != nil {
			return fmt.Errorf("insert release transaction: %w", err)
		}

		w.PendingBalance = w.PendingBalance.Sub(amount)
		if w.PendingBalance.IsNegative() {
			w.PendingBalance = decimal.Zero
		}
		w.AvailableBalance = w.AvailableBalance.Add(amount)
		w.LastTransactionID = &tx.ID
		return wallets.UpdateBalances(ctx, w)
	}

	if s.pool != nil {
		dbTx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		defer dbTx.Rollback(ctx) //nolint:errcheck

		if err := doRelease(s.wallets.WithTx(dbTx), s.transactions.WithTx(dbTx)); err != nil {
			return err
		}
		return dbTx.Commit(ctx)
	}
	return doRelease(s.wallets, s.transactions)
}

// VerifyChain re-computes every HMAC for the wallet and returns ErrChainTampered
// if any entry does not match. On mismatch it also suspends the tenant.
func (s *WalletService) VerifyChain(ctx context.Context, walletID uuid.UUID, tenantID uuid.UUID) error {
	const pageSize = 500
	offset := 0
	var prevSig string

	for {
		txs, err := s.transactions.ListByWalletAsc(ctx, walletID, pageSize, offset)
		if err != nil {
			return fmt.Errorf("list transactions: %w", err)
		}
		if len(txs) == 0 {
			break
		}

		for _, tx := range txs {
			expected := computeSignature(tx.Amount, tx.RunningBalance, prevSig, s.secret)
			if !hmac.Equal([]byte(tx.Signature), []byte(expected)) {
				_ = s.suspendTenant(ctx, tenantID)
				_ = s.logChainBreach(ctx, tenantID, walletID)
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
// toPending controls whether the credit goes to pending_balance (true) or available_balance (false).
// When a pool is configured, the entire operation runs inside a DB transaction with
// SELECT ... FOR UPDATE on the wallet row to prevent concurrent balance corruption.
func (s *WalletService) record(ctx context.Context, tenantID uuid.UUID, amount decimal.Decimal, txType models.TransactionType, toPending bool, orderID *uuid.UUID) (*models.Transaction, error) {
	if s.pool != nil {
		return s.recordTx(ctx, tenantID, amount, txType, toPending, orderID)
	}
	return s.doRecord(ctx, s.wallets, s.transactions, tenantID, amount, txType, toPending, orderID)
}

func (s *WalletService) recordTx(ctx context.Context, tenantID uuid.UUID, amount decimal.Decimal, txType models.TransactionType, toPending bool, orderID *uuid.UUID) (*models.Transaction, error) {
	dbTx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer dbTx.Rollback(ctx) //nolint:errcheck

	wallets := s.wallets.WithTx(dbTx)
	transactions := s.transactions.WithTx(dbTx)

	result, err := s.doRecordForUpdate(ctx, wallets, transactions, tenantID, amount, txType, toPending, orderID)
	if err != nil {
		return nil, err
	}

	if err := dbTx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return result, nil
}

func (s *WalletService) doRecordForUpdate(ctx context.Context, wallets repository.WalletRepository, transactions repository.TransactionRepository, tenantID uuid.UUID, amount decimal.Decimal, txType models.TransactionType, toPending bool, orderID *uuid.UUID) (*models.Transaction, error) {
	w, err := wallets.GetByTenantIDForUpdate(ctx, tenantID)
	if err != nil {
		return nil, ErrWalletNotFound
	}
	return s.doRecordWithWallet(ctx, wallets, transactions, w, amount, txType, toPending, orderID)
}

func (s *WalletService) doRecord(ctx context.Context, wallets repository.WalletRepository, transactions repository.TransactionRepository, tenantID uuid.UUID, amount decimal.Decimal, txType models.TransactionType, toPending bool, orderID *uuid.UUID) (*models.Transaction, error) {
	w, err := wallets.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, ErrWalletNotFound
	}
	return s.doRecordWithWallet(ctx, wallets, transactions, w, amount, txType, toPending, orderID)
}

func (s *WalletService) doRecordWithWallet(ctx context.Context, wallets repository.WalletRepository, transactions repository.TransactionRepository, w *models.Wallet, amount decimal.Decimal, txType models.TransactionType, toPending bool, orderID *uuid.UUID) (*models.Transaction, error) {
	var prevSig string
	var prevRunningBalance decimal.Decimal
	if w.LastTransactionID != nil {
		prev, err := transactions.GetLatestByWallet(ctx, w.ID)
		if err != nil {
			return nil, fmt.Errorf("fetch chain tip: %w", err)
		}
		prevSig = prev.Signature
		prevRunningBalance = prev.RunningBalance
	}

	newRunningBalance := prevRunningBalance.Add(amount)
	sig := computeSignature(amount, newRunningBalance, prevSig, s.secret)

	tx := &models.Transaction{
		WalletID:       w.ID,
		OrderID:        orderID,
		Amount:         amount,
		RunningBalance: newRunningBalance,
		Type:           txType,
		Signature:      sig,
	}
	if err := transactions.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	if toPending {
		w.PendingBalance = w.PendingBalance.Add(amount)
	} else {
		w.AvailableBalance = w.AvailableBalance.Add(amount)
	}
	w.LastTransactionID = &tx.ID
	if err := wallets.UpdateBalances(ctx, w); err != nil {
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

func (s *WalletService) logChainBreach(ctx context.Context, tenantID, walletID uuid.UUID) error {
	if s.auditLogs == nil {
		return nil
	}
	diff := fmt.Sprintf(`{"wallet_id":"%s","reason":"HMAC chain tampered"}`, walletID)
	return s.auditLogs.Create(ctx, &models.AuditLog{
		TenantID: tenantID,
		Action:   "chain_tampered",
		Diff:     json.RawMessage(diff),
	})
}

// computeSignature returns HMAC-SHA256(amount|running_balance|prev_sig, secret).
func computeSignature(amount, runningBalance decimal.Decimal, prevSig, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(amount.String()))
	mac.Write([]byte(runningBalance.String()))
	mac.Write([]byte(prevSig))
	return hex.EncodeToString(mac.Sum(nil))
}

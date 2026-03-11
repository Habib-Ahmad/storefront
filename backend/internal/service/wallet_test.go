package service_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

const testHMACSecret = "test-hmac-secret-32-bytes-minimum!"

func sign(amount, balance decimal.Decimal, prev, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(amount.String()))
	mac.Write([]byte(balance.String()))
	mac.Write([]byte(prev))
	return hex.EncodeToString(mac.Sum(nil))
}

func newWalletSvc(w *models.Wallet, txRepo *mockTxRepo, tenantRepo *mockTenantRepo) *service.WalletService {
	return service.NewWalletService(
		&mockWalletRepo{wallet: w},
		txRepo,
		tenantRepo,
		testHMACSecret,
	)
}

func TestCredit_FirstEntry_NoChain(t *testing.T) {
	walletID := uuid.New()
	w := &models.Wallet{ID: walletID, TenantID: uuid.New()}
	txRepo := &mockTxRepo{}
	svc := newWalletSvc(w, txRepo, &mockTenantRepo{})

	amount := decimal.NewFromInt(1000)
	tx, err := svc.Credit(context.Background(), walletID, amount, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedSig := sign(amount, amount, "", testHMACSecret)
	if tx.Signature != expectedSig {
		t.Fatalf("signature mismatch\ngot:  %s\nwant: %s", tx.Signature, expectedSig)
	}
	if !tx.RunningBalance.Equal(amount) {
		t.Fatalf("expected running_balance=%s, got %s", amount, tx.RunningBalance)
	}
}

func TestCredit_ChainContinues(t *testing.T) {
	walletID := uuid.New()
	prevTxID := uuid.New()
	prevSig := sign(decimal.NewFromInt(1000), decimal.NewFromInt(1000), "", testHMACSecret)

	w := &models.Wallet{
		ID:                walletID,
		TenantID:          uuid.New(),
		AvailableBalance:  decimal.NewFromInt(1000),
		LastTransactionID: &prevTxID,
	}
	txRepo := &mockTxRepo{latest: &models.Transaction{ID: prevTxID, Signature: prevSig}}
	svc := newWalletSvc(w, txRepo, &mockTenantRepo{})

	amount := decimal.NewFromInt(500)
	tx, err := svc.Credit(context.Background(), walletID, amount, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedBalance := decimal.NewFromInt(1500)
	expectedSig := sign(amount, expectedBalance, prevSig, testHMACSecret)
	if tx.Signature != expectedSig {
		t.Fatalf("chain signature mismatch\ngot:  %s\nwant: %s", tx.Signature, expectedSig)
	}
}

func TestVerifyChain_ValidChain(t *testing.T) {
	walletID := uuid.New()
	tenantID := uuid.New()
	w := &models.Wallet{ID: walletID, TenantID: tenantID}

	amount1 := decimal.NewFromInt(1000)
	bal1 := decimal.NewFromInt(1000)
	sig1 := sign(amount1, bal1, "", testHMACSecret)

	amount2 := decimal.NewFromInt(500)
	bal2 := decimal.NewFromInt(1500)
	sig2 := sign(amount2, bal2, sig1, testHMACSecret)

	// ListByWallet returns newest-first; VerifyChain reverses internally.
	txs := []models.Transaction{
		{Amount: amount2, RunningBalance: bal2, Signature: sig2},
		{Amount: amount1, RunningBalance: bal1, Signature: sig1},
	}
	txRepo := &mockTxRepo{txs: txs}
	svc := service.NewWalletService(&mockWalletRepo{wallet: w}, txRepo, &mockTenantRepo{}, testHMACSecret)

	if err := svc.VerifyChain(context.Background(), walletID, tenantID); err != nil {
		t.Fatalf("expected valid chain, got: %v", err)
	}
}

func TestVerifyChain_Tampered_SuspendsTenant(t *testing.T) {
	walletID := uuid.New()
	tenantID := uuid.New()
	w := &models.Wallet{ID: walletID, TenantID: tenantID}

	tamperedTx := models.Transaction{
		Amount:         decimal.NewFromInt(1000),
		RunningBalance: decimal.NewFromInt(1000),
		Signature:      "tampered-signature",
	}
	txRepo := &mockTxRepo{txs: []models.Transaction{tamperedTx}}
	tenantRepo := &mockTenantRepo{tenant: &models.Tenant{ID: tenantID, Status: models.TenantStatusActive}}
	svc := service.NewWalletService(&mockWalletRepo{wallet: w}, txRepo, tenantRepo, testHMACSecret)

	err := svc.VerifyChain(context.Background(), walletID, tenantID)
	if err == nil {
		t.Fatal("expected ErrChainTampered")
	}
	if tenantRepo.updated == nil || tenantRepo.updated.Status != models.TenantStatusSuspended {
		t.Fatal("tenant should be suspended on chain violation")
	}
}

func TestDebit_ChainEntry(t *testing.T) {
	// Debit must produce a signed negative-amount ledger entry that chains correctly.
	walletID := uuid.New()
	w := &models.Wallet{ID: walletID, TenantID: uuid.New(), AvailableBalance: decimal.NewFromInt(1000)}
	txRepo := &mockTxRepo{}
	svc := newWalletSvc(w, txRepo, &mockTenantRepo{})

	debitAmount := decimal.NewFromInt(300)
	tx, err := svc.Debit(context.Background(), walletID, debitAmount, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedBalance := decimal.NewFromInt(700) // 1000 - 300
	if !tx.RunningBalance.Equal(expectedBalance) {
		t.Fatalf("running_balance: want 700, got %s", tx.RunningBalance)
	}
	// Debit records the negated amount in the ledger; signature must use that negated value.
	negatedAmount := debitAmount.Neg()
	expectedSig := sign(negatedAmount, expectedBalance, "", testHMACSecret)
	if tx.Signature != expectedSig {
		t.Fatalf("debit signature mismatch\ngot:  %s\nwant: %s", tx.Signature, expectedSig)
	}
}

func TestReleasePending_MovesBalance(t *testing.T) {
	// Spec §7: PENDING_BALANCE moves to AVAILABLE_BALANCE only on Terminal Africa 'delivered' webhook.
	tenantID := uuid.New()
	w := &models.Wallet{
		ID:               uuid.New(),
		TenantID:         tenantID,
		PendingBalance:   decimal.NewFromInt(1000),
		AvailableBalance: decimal.NewFromInt(500),
	}
	walletRepo := &mockWalletRepo{wallet: w}
	svc := service.NewWalletService(walletRepo, &mockTxRepo{}, &mockTenantRepo{}, testHMACSecret)

	if err := svc.ReleasePending(context.Background(), tenantID, decimal.NewFromInt(1000)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if walletRepo.updated == nil {
		t.Fatal("wallet balances were not updated")
	}
	if !walletRepo.updated.AvailableBalance.Equal(decimal.NewFromInt(1500)) {
		t.Fatalf("available_balance: want 1500, got %s", walletRepo.updated.AvailableBalance)
	}
	if !walletRepo.updated.PendingBalance.Equal(decimal.NewFromInt(0)) {
		t.Fatalf("pending_balance: want 0, got %s", walletRepo.updated.PendingBalance)
	}
}

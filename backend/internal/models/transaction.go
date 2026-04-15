package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	TransactionTypeCredit     TransactionType = "credit"
	TransactionTypeDebit      TransactionType = "debit"
	TransactionTypeCommission TransactionType = "commission"
	TransactionTypePayout     TransactionType = "payout"
	TransactionTypeRelease    TransactionType = "release"
	TransactionTypeRefund     TransactionType = "refund"
)

// Insert-only. DB Rules block UPDATE and DELETE.
// Signature = HMAC-SHA256(amount + running_balance + prev_signature + secret).
type Transaction struct {
	ID                uuid.UUID       `json:"id"`
	WalletID          uuid.UUID       `json:"wallet_id"`
	OrderID           *uuid.UUID      `json:"order_id,omitempty"`
	Amount            decimal.Decimal `json:"amount"`
	RunningBalance    decimal.Decimal `json:"running_balance"`
	PlatformFeeBase   decimal.Decimal `json:"platform_fee_base"`
	PlatformFeeRate   decimal.Decimal `json:"platform_fee_rate"`
	PlatformFeeCap    decimal.Decimal `json:"platform_fee_cap"`
	PlatformFeeAmount decimal.Decimal `json:"platform_fee_amount"`
	Type              TransactionType `json:"type"`
	Signature         string          `json:"signature"`
	CreatedAt         time.Time       `json:"created_at"`
}

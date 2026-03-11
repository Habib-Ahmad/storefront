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
)

// Insert-only. DB Rules block UPDATE and DELETE.
// Signature = HMAC-SHA256(amount + running_balance + prev_signature + secret).
type Transaction struct {
	ID             uuid.UUID
	WalletID       uuid.UUID
	OrderID        *uuid.UUID // nil for manual adjustments
	Amount         decimal.Decimal
	RunningBalance decimal.Decimal
	Type           TransactionType
	Signature      string
	CreatedAt      time.Time
}

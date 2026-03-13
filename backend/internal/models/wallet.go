package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// No FK on LastTransactionID — avoids circular reference with transactions.
type Wallet struct {
	ID                   uuid.UUID       `json:"id"`
	TenantID             uuid.UUID       `json:"tenant_id"`
	AvailableBalance     decimal.Decimal `json:"available_balance"`
	PendingBalance       decimal.Decimal `json:"pending_balance"`
	LastTransactionID    *uuid.UUID      `json:"last_transaction_id,omitempty"`
	LastReconciliationAt *time.Time      `json:"last_reconciliation_at,omitempty"`
}

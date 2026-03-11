package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// No FK on LastTransactionID — avoids circular reference with transactions.
type Wallet struct {
	ID                   uuid.UUID
	TenantID             uuid.UUID
	AvailableBalance     decimal.Decimal
	PendingBalance       decimal.Decimal
	LastTransactionID    *uuid.UUID
	LastReconciliationAt *time.Time
}

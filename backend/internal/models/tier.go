package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Tier struct {
	ID             uuid.UUID
	Name           string
	DebtCeiling    decimal.Decimal
	CommissionRate decimal.Decimal
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

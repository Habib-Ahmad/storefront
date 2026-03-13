package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Tier struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	DebtCeiling    decimal.Decimal `json:"debt_ceiling"`
	CommissionRate decimal.Decimal `json:"commission_rate"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

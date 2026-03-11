package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ProductVariant struct {
	ID         uuid.UUID
	ProductID  uuid.UUID
	SKU        string
	Attributes json.RawMessage
	Price      decimal.Decimal
	StockQty   *int // nil = infinite, 0 = sold out
	IsDefault  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

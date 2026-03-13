package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ProductVariant struct {
	ID         uuid.UUID       `json:"id"`
	ProductID  uuid.UUID       `json:"product_id"`
	SKU        string          `json:"sku"`
	Attributes json.RawMessage `json:"attributes"`
	Price      decimal.Decimal `json:"price"`
	StockQty   *int            `json:"stock_qty,omitempty"` // nil = infinite, 0 = sold out
	IsDefault  bool            `json:"is_default"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	DeletedAt  *time.Time      `json:"-"`
}

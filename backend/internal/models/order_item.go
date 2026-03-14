package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderItem struct {
	ID              uuid.UUID        `json:"id"`
	OrderID         uuid.UUID        `json:"order_id"`
	VariantID       uuid.UUID        `json:"variant_id"`
	Quantity        int              `json:"quantity"`
	PriceAtSale     decimal.Decimal  `json:"price_at_sale"`
	CostPriceAtSale *decimal.Decimal `json:"cost_price_at_sale,omitempty"`
	ProductName     *string          `json:"product_name,omitempty"`
	VariantLabel    *string          `json:"variant_label,omitempty"`
}

package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderItem struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	VariantID   uuid.UUID
	Quantity    int
	PriceAtSale decimal.Decimal
}

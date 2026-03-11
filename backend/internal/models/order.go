package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusFailed  PaymentStatus = "failed"
)

type FulfillmentStatus string

const (
	FulfillmentStatusProcessing FulfillmentStatus = "processing"
	FulfillmentStatusShipped    FulfillmentStatus = "shipped"
	FulfillmentStatusDelivered  FulfillmentStatus = "delivered"
)

type Order struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	TrackingSlug      string
	IsDelivery        bool
	CustomerName      string
	CustomerPhone     *string
	CustomerEmail     *string
	ShippingAddress   *string
	TotalAmount       decimal.Decimal
	ShippingFee       decimal.Decimal
	PaymentStatus     PaymentStatus
	FulfillmentStatus FulfillmentStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentMethod string

const (
	PaymentMethodOnline   PaymentMethod = "online"
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodTransfer PaymentMethod = "transfer"
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
	ID                uuid.UUID         `json:"id"`
	TenantID          uuid.UUID         `json:"tenant_id"`
	TrackingSlug      string            `json:"tracking_slug"`
	IsDelivery        bool              `json:"is_delivery"`
	CustomerName      *string           `json:"customer_name,omitempty"`
	CustomerPhone     *string           `json:"customer_phone,omitempty"`
	CustomerEmail     *string           `json:"customer_email,omitempty"`
	ShippingAddress   *string           `json:"shipping_address,omitempty"`
	TotalAmount       decimal.Decimal   `json:"total_amount"`
	ShippingFee       decimal.Decimal   `json:"shipping_fee"`
	PaymentMethod     PaymentMethod     `json:"payment_method"`
	PaymentStatus     PaymentStatus     `json:"payment_status"`
	FulfillmentStatus FulfillmentStatus `json:"fulfillment_status"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

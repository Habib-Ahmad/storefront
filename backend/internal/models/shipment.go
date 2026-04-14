package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ShipmentStatus string

const (
	ShipmentStatusQueued    ShipmentStatus = "queued"
	ShipmentStatusPickedUp  ShipmentStatus = "picked_up"
	ShipmentStatusInTransit ShipmentStatus = "in_transit"
	ShipmentStatusDelivered ShipmentStatus = "delivered"
	ShipmentStatusFailed    ShipmentStatus = "failed"
)

type Shipment struct {
	ID             uuid.UUID       `json:"id"`
	OrderID        uuid.UUID       `json:"order_id"`
	TenantID       uuid.UUID       `json:"tenant_id"`
	Status         ShipmentStatus  `json:"status"`
	CarrierRef     *string         `json:"carrier_ref,omitempty"`
	TrackingNumber *string         `json:"tracking_number,omitempty"`
	CarrierHistory json.RawMessage `json:"carrier_history,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type DispatchShipmentOption struct {
	ID            string          `json:"id"`
	CourierID     string          `json:"courier_id"`
	CourierName   string          `json:"courier_name"`
	ServiceCode   string          `json:"service_code"`
	ServiceType   string          `json:"service_type"`
	Amount        string          `json:"amount"`
	Currency      string          `json:"currency"`
	PickupETA     string          `json:"pickup_eta,omitempty"`
	DeliveryETA   string          `json:"delivery_eta,omitempty"`
	TrackingLabel string          `json:"tracking_label,omitempty"`
	TrackingLevel int             `json:"tracking_level"`
	IsFastest     bool            `json:"is_fastest"`
	IsCheapest    bool            `json:"is_cheapest"`
	ProviderData  json.RawMessage `json:"provider_data,omitempty"`
}

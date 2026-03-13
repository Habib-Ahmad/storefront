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

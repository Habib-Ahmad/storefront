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
	ID             uuid.UUID
	OrderID        uuid.UUID
	TenantID       uuid.UUID
	Status         ShipmentStatus
	CarrierRef     *string // external carrier booking ID
	TrackingNumber *string
	CarrierHistory json.RawMessage // raw logs from Terminal Africa / Shipbubble
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

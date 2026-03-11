package models

import (
	"time"

	"github.com/google/uuid"
)

type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
)

type ActiveModules struct {
	Inventory bool `json:"inventory"`
	Payments  bool `json:"payments"`
	Logistics bool `json:"logistics"`
}

type Tenant struct {
	ID                   uuid.UUID
	TierID               uuid.UUID
	Name                 string
	Slug                 string
	PaystackSubaccountID *string
	ActiveModules        ActiveModules
	Status               TenantStatus
	CreatedAt            time.Time
	UpdatedAt            time.Time
	DeletedAt            *time.Time
}

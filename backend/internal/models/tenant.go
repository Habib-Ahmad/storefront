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
	ID                   uuid.UUID     `json:"id"`
	TierID               uuid.UUID     `json:"tier_id"`
	Name                 string        `json:"name"`
	Slug                 string        `json:"slug"`
	PaystackSubaccountID *string       `json:"paystack_subaccount_id,omitempty"`
	ActiveModules        ActiveModules `json:"active_modules"`
	Status               TenantStatus  `json:"status"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
	DeletedAt            *time.Time    `json:"-"`
}

package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID        `json:"id"`
	TenantID    uuid.UUID        `json:"tenant_id"`
	Name        string           `json:"name"`
	Description *string          `json:"description,omitempty"`
	Category    *string          `json:"category,omitempty"`
	IsAvailable bool             `json:"is_available"`
	Variants    []ProductVariant `json:"variants,omitempty"`
	Images      []ProductImage   `json:"images,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	DeletedAt   *time.Time       `json:"-"`
}

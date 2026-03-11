package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description *string
	Category    *string
	IsAvailable bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

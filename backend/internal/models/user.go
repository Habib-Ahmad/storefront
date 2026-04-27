package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleStaff UserRole = "staff"
)

type User struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	Email     string     `json:"email"`
	FirstName *string    `json:"first_name,omitempty"`
	LastName  *string    `json:"last_name,omitempty"`
	Phone     *string    `json:"phone,omitempty"`
	Role      UserRole   `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

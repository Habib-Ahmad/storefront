package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	UserRoleAdmin   UserRole = "admin"
	UserRoleStaff   UserRole = "staff"
	UserRoleManager UserRole = "manager"
)

type User struct {
	ID        uuid.UUID  `json:"id"` // set by Supabase Auth, not DB-generated
	TenantID  uuid.UUID  `json:"tenant_id"`
	Email     string     `json:"email"`
	Role      UserRole   `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

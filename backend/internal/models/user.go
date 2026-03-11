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
	ID        uuid.UUID // set by Supabase Auth, not DB-generated
	TenantID  uuid.UUID
	Email     string
	Role      UserRole
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

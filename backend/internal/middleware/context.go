package middleware

import (
	"context"

	"github.com/google/uuid"

	"storefront/backend/internal/models"
)

type contextKey int

const (
	ctxKeyUserID contextKey = iota
	ctxKeyUserRole
	ctxKeyTenant
)

func UserIDFromCtx(ctx context.Context) uuid.UUID {
	v, _ := ctx.Value(ctxKeyUserID).(uuid.UUID)
	return v
}

func UserRoleFromCtx(ctx context.Context) models.UserRole {
	v, _ := ctx.Value(ctxKeyUserRole).(models.UserRole)
	return v
}

func TenantFromCtx(ctx context.Context) *models.Tenant {
	v, _ := ctx.Value(ctxKeyTenant).(*models.Tenant)
	return v
}

func setTenant(ctx context.Context, t *models.Tenant) context.Context {
	return context.WithValue(ctx, ctxKeyTenant, t)
}

// WithUserID injects a user ID into ctx. Intended for use in tests.
func WithUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxKeyUserID, id)
}

// WithTenant injects a tenant into ctx. Intended for use in tests.
func WithTenant(ctx context.Context, t *models.Tenant) context.Context {
	return setTenant(ctx, t)
}

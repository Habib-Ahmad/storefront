package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

var (
	ErrTenantNotFound = errors.New("tenant not found")
	ErrModuleDisabled = errors.New("module not enabled for this tenant")
)

type TenantService struct {
	tenants repository.TenantRepository
	wallets repository.WalletRepository
	users   repository.UserRepository
}

func NewTenantService(
	tenants repository.TenantRepository,
	wallets repository.WalletRepository,
	users repository.UserRepository,
) *TenantService {
	return &TenantService{tenants: tenants, wallets: wallets, users: users}
}

// Onboard creates a new tenant, its first admin user, and an empty wallet.
func (s *TenantService) Onboard(ctx context.Context, tenantName, slug string, tierID, adminUserID uuid.UUID, adminEmail string) (*models.Tenant, error) {
	tenant := &models.Tenant{
		TierID: tierID,
		Name:   tenantName,
		Slug:   slug,
		Status: models.TenantStatusActive,
	}
	if err := s.tenants.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	user := &models.User{
		ID:       adminUserID,
		TenantID: tenant.ID,
		Email:    adminEmail,
		Role:     models.UserRoleAdmin,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create admin user: %w", err)
	}

	wallet := &models.Wallet{TenantID: tenant.ID}
	if err := s.wallets.Create(ctx, wallet); err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}

	return tenant, nil
}

// SetModules replaces the tenant's active_modules configuration.
func (s *TenantService) SetModules(ctx context.Context, tenantID uuid.UUID, modules models.ActiveModules) error {
	tenant, err := s.tenants.GetByID(ctx, tenantID)
	if err != nil {
		return ErrTenantNotFound
	}
	tenant.ActiveModules = modules
	return s.tenants.Update(ctx, tenant)
}

// RequireModule returns ErrModuleDisabled if the requested module is not active.
func RequireModule(tenant *models.Tenant, inventory, payments, logistics bool) error {
	if inventory && !tenant.ActiveModules.Inventory {
		return ErrModuleDisabled
	}
	if payments && !tenant.ActiveModules.Payments {
		return ErrModuleDisabled
	}
	if logistics && !tenant.ActiveModules.Logistics {
		return ErrModuleDisabled
	}
	return nil
}

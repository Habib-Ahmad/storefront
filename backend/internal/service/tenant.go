package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

const defaultTierName = "Standard"

var (
	ErrTenantNotFound = errors.New("tenant not found")
	ErrModuleDisabled = errors.New("module not enabled for this tenant")
	ErrSlugTaken      = errors.New("slug already in use")
	ErrUserExists     = errors.New("user already belongs to a tenant")
)

type TenantService struct {
	tenants repository.TenantRepository
	tiers   repository.TierRepository
	wallets repository.WalletRepository
	users   repository.UserRepository
}

func NewTenantService(
	tenants repository.TenantRepository,
	tiers repository.TierRepository,
	wallets repository.WalletRepository,
	users repository.UserRepository,
) *TenantService {
	return &TenantService{tenants: tenants, tiers: tiers, wallets: wallets, users: users}
}

// Onboard creates a new tenant, its first admin user, and an empty wallet.
// Every new vendor is assigned the Standard tier automatically.
func (s *TenantService) Onboard(ctx context.Context, tenantName, slug string, adminUserID uuid.UUID, adminEmail string) (*models.Tenant, error) {
	tier, err := s.tiers.GetByName(ctx, defaultTierName)
	if err != nil {
		return nil, fmt.Errorf("lookup default tier: %w", err)
	}

	tenant := &models.Tenant{
		TierID:       tier.ID,
		Name:         tenantName,
		Slug:         slug,
		ContactEmail: &adminEmail,
		Status:       models.TenantStatusActive,
	}
	if err := s.tenants.Create(ctx, tenant); err != nil {
		if strings.Contains(err.Error(), "tenants_slug_key") {
			return nil, ErrSlugTaken
		}
		return nil, fmt.Errorf("create tenant: %w", err)
	}

	user := &models.User{
		ID:       adminUserID,
		TenantID: tenant.ID,
		Email:    adminEmail,
		Role:     models.UserRoleAdmin,
	}
	if err := s.users.Create(ctx, user); err != nil {
		if strings.Contains(err.Error(), "users_pkey") {
			return nil, ErrUserExists
		}
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

// UpdateProfile updates the tenant's editable profile fields.
func (s *TenantService) UpdateProfile(ctx context.Context, tenantID uuid.UUID, name string, contactEmail, contactPhone, address, logoURL *string) error {
	tenant, err := s.tenants.GetByID(ctx, tenantID)
	if err != nil {
		return ErrTenantNotFound
	}
	tenant.Name = name
	tenant.ContactEmail = contactEmail
	tenant.ContactPhone = contactPhone
	tenant.Address = address
	tenant.LogoURL = logoURL
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

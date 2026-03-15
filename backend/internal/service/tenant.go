package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"storefront/backend/internal/apperr"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

const defaultTierName = "Standard"

var (
	ErrTenantNotFound = apperr.NotFound("tenant not found")
	ErrModuleDisabled = apperr.Forbidden("module not enabled for this tenant")
	ErrSlugTaken      = apperr.Conflict("slug already in use")
	ErrUserExists     = apperr.Conflict("user already belongs to a tenant")
)

type TenantService struct {
	tenants repository.TenantRepository
	tiers   repository.TierRepository
	wallets repository.WalletRepository
	users   repository.UserRepository
	pool    TxBeginner
}

func NewTenantService(
	tenants repository.TenantRepository,
	tiers repository.TierRepository,
	wallets repository.WalletRepository,
	users repository.UserRepository,
) *TenantService {
	return &TenantService{tenants: tenants, tiers: tiers, wallets: wallets, users: users}
}

func (s *TenantService) SetPool(pool TxBeginner) { s.pool = pool }

// Onboard creates a new tenant, its first admin user, and an empty wallet.
// Every new vendor is assigned the Standard tier automatically.
// All three inserts run in a single DB transaction when a pool is configured.
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
	user := &models.User{
		ID:       adminUserID,
		TenantID: uuid.Nil, // set after tenant create
		Email:    adminEmail,
		Role:     models.UserRoleAdmin,
	}

	create := func(tenants repository.TenantRepository, users repository.UserRepository, wallets repository.WalletRepository) error {
		if err := tenants.Create(ctx, tenant); err != nil {
			if apperr.IsUniqueViolation(err) {
				return ErrSlugTaken
			}
			return fmt.Errorf("create tenant: %w", err)
		}
		user.TenantID = tenant.ID
		if err := users.Create(ctx, user); err != nil {
			if apperr.IsUniqueViolation(err) {
				return ErrUserExists
			}
			return fmt.Errorf("create admin user: %w", err)
		}
		wallet := &models.Wallet{TenantID: tenant.ID}
		if err := wallets.Create(ctx, wallet); err != nil {
			return fmt.Errorf("create wallet: %w", err)
		}
		return nil
	}

	if s.pool != nil {
		dbTx, err := s.pool.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("begin tx: %w", err)
		}
		defer dbTx.Rollback(ctx) //nolint:errcheck

		if err := create(s.tenants.WithTx(dbTx), s.users.WithTx(dbTx), s.wallets.WithTx(dbTx)); err != nil {
			return nil, err
		}
		if err := dbTx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit onboard tx: %w", err)
		}
		return tenant, nil
	}

	// Fallback: non-transactional path for tests / simple setups.
	if err := create(s.tenants, s.users, s.wallets); err != nil {
		return nil, err
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

package service

import (
	"context"
	"fmt"
	"strings"

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

const (
	defaultStorefrontSlug = "store"
	storefrontSlugMaxLen  = 50
	maxSlugAttempts       = 5
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
func (s *TenantService) Onboard(ctx context.Context, tenantName string, adminUserID uuid.UUID, adminEmail string) (*models.Tenant, error) {
	tier, err := s.tiers.GetByName(ctx, defaultTierName)
	if err != nil {
		return nil, fmt.Errorf("lookup default tier: %w", err)
	}

	tenant := &models.Tenant{
		TierID:              tier.ID,
		Name:                tenantName,
		StorefrontPublished: false,
		ContactEmail:        &adminEmail,
		Status:              models.TenantStatusActive,
	}
	user := &models.User{
		ID:       adminUserID,
		TenantID: uuid.Nil,
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

	createWithSlugRetries := func(tenants repository.TenantRepository, users repository.UserRepository, wallets repository.WalletRepository) error {
		for attempt := range maxSlugAttempts {
			tenant.Slug = generateTemporaryStorefrontSlug(tenantName, attempt)
			err := create(tenants, users, wallets)
			if err == nil {
				return nil
			}
			if err != ErrSlugTaken {
				return err
			}
		}
		return ErrSlugTaken
	}

	if s.pool != nil {
		dbTx, err := s.pool.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("begin tx: %w", err)
		}
		defer dbTx.Rollback(ctx) //nolint:errcheck

		if err := createWithSlugRetries(s.tenants.WithTx(dbTx), s.users.WithTx(dbTx), s.wallets.WithTx(dbTx)); err != nil {
			return nil, err
		}
		if err := dbTx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("commit onboard tx: %w", err)
		}
		return tenant, nil
	}

	if err := createWithSlugRetries(s.tenants, s.users, s.wallets); err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *TenantService) UpdateStorefront(ctx context.Context, tenantID uuid.UUID, slug string, published bool) error {
	tenant, err := s.tenants.GetByID(ctx, tenantID)
	if err != nil {
		return ErrTenantNotFound
	}

	tenant.Slug = slug
	tenant.StorefrontPublished = published

	if err := s.tenants.Update(ctx, tenant); err != nil {
		if apperr.IsUniqueViolation(err) {
			return ErrSlugTaken
		}
		return err
	}

	return nil
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

func generateTemporaryStorefrontSlug(name string, attempt int) string {
	base := slugifyStorefrontName(name)
	if base == "" {
		base = defaultStorefrontSlug
	}

	suffix := ""
	if attempt > 0 {
		suffix = "-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:4]
	}

	maxBaseLen := storefrontSlugMaxLen - len(suffix)
	if maxBaseLen < len(defaultStorefrontSlug) {
		maxBaseLen = len(defaultStorefrontSlug)
	}
	if len(base) > maxBaseLen {
		base = strings.Trim(base[:maxBaseLen], "-")
		if base == "" {
			base = defaultStorefrontSlug
		}
	}

	return base + suffix
}

func slugifyStorefrontName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}

	var builder strings.Builder
	builder.Grow(len(value))
	lastHyphen := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastHyphen = false
		case !lastHyphen && builder.Len() > 0:
			builder.WriteByte('-')
			lastHyphen = true
		}
	}

	return strings.Trim(builder.String(), "-")
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

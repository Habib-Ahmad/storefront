package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

func TestOnboard_CreatesTenanUserAndWallet(t *testing.T) {
	tenantRepo := &mockTenantRepo{}
	walletRepo := &mockWalletRepo{}
	userRepo := &mockUserRepo{}

	svc := service.NewTenantService(tenantRepo, walletRepo, userRepo)
	tenant, err := svc.Onboard(context.Background(), "Acme", "acme", uuid.New(), uuid.New(), "admin@acme.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenant == nil || tenant.Name != "Acme" {
		t.Fatal("tenant not returned correctly")
	}
	if userRepo.user == nil || userRepo.user.Role != models.UserRoleAdmin {
		t.Fatal("admin user not created")
	}
	if walletRepo.wallet == nil {
		t.Fatal("wallet not created")
	}
}

func TestSetModules_UpdatesTenant(t *testing.T) {
	tenantID := uuid.New()
	tenantRepo := &mockTenantRepo{tenant: &models.Tenant{ID: tenantID, Status: models.TenantStatusActive}}
	svc := service.NewTenantService(tenantRepo, &mockWalletRepo{}, &mockUserRepo{})

	mods := models.ActiveModules{Inventory: true, Payments: true}
	if err := svc.SetModules(context.Background(), tenantID, mods); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenantRepo.updated == nil {
		t.Fatal("tenant not updated")
	}
	if !tenantRepo.updated.ActiveModules.Inventory {
		t.Fatal("inventory module not set")
	}
}

func TestRequireModule_RejectsDisabled(t *testing.T) {
	tenant := &models.Tenant{ActiveModules: models.ActiveModules{Inventory: false}}
	if err := service.RequireModule(tenant, true, false, false); err == nil {
		t.Fatal("expected error for disabled inventory module")
	}
}

func TestRequireModule_PassesEnabled(t *testing.T) {
	tenant := &models.Tenant{ActiveModules: models.ActiveModules{Inventory: true, Payments: true}}
	if err := service.RequireModule(tenant, true, true, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOnboard_AdminUserBelongsToTenant(t *testing.T) {
	// The admin user's TenantID must be set to the newly created tenant's ID.
	tenantRepo := &mockTenantRepo{}
	userRepo := &mockUserRepo{}
	svc := service.NewTenantService(tenantRepo, &mockWalletRepo{}, userRepo)

	tenant, err := svc.Onboard(context.Background(), "Acme", "acme", uuid.New(), uuid.New(), "admin@acme.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userRepo.user == nil {
		t.Fatal("admin user not created")
	}
	if userRepo.user.TenantID != tenant.ID {
		t.Fatalf("admin user TenantID: want %s, got %s", tenant.ID, userRepo.user.TenantID)
	}
}

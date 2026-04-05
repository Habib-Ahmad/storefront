package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"storefront/backend/internal/db"
	"storefront/backend/internal/handler"
	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
	"storefront/backend/internal/service"
)

// ── minimal mock repos for TenantService ─────────────────────────────────────

type stubTenantRepo struct {
	tenant    *models.Tenant
	createErr error
	updateErr error
}

func (s *stubTenantRepo) Create(_ context.Context, t *models.Tenant) error {
	if s.createErr != nil {
		return s.createErr
	}
	t.ID = uuid.New()
	s.tenant = t
	return nil
}
func (s *stubTenantRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Tenant, error) {
	return s.tenant, nil
}
func (s *stubTenantRepo) GetBySlug(_ context.Context, _ string) (*models.Tenant, error) {
	return s.tenant, nil
}
func (s *stubTenantRepo) Update(_ context.Context, t *models.Tenant) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.tenant = t
	return nil
}
func (s *stubTenantRepo) SoftDelete(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubTenantRepo) WithTx(_ db.DBTX) repository.TenantRepository    { return s }

type stubWalletRepo struct{}

func (s *stubWalletRepo) Create(_ context.Context, w *models.Wallet) error {
	w.ID = uuid.New()
	return nil
}
func (s *stubWalletRepo) GetByTenantID(_ context.Context, _ uuid.UUID) (*models.Wallet, error) {
	return &models.Wallet{}, nil
}
func (s *stubWalletRepo) GetByTenantIDForUpdate(_ context.Context, _ uuid.UUID) (*models.Wallet, error) {
	return &models.Wallet{}, nil
}
func (s *stubWalletRepo) UpdateBalances(_ context.Context, _ *models.Wallet) error { return nil }
func (s *stubWalletRepo) WithTx(_ db.DBTX) repository.WalletRepository             { return s }
func (s *stubWalletRepo) ListActiveWallets(_ context.Context) ([]repository.ActiveWallet, error) {
	return nil, nil
}

type stubUserRepo struct{}

func (s *stubUserRepo) Create(_ context.Context, _ *models.User) error { return nil }
func (s *stubUserRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.User, error) {
	return nil, nil
}
func (s *stubUserRepo) GetByEmail(_ context.Context, _ uuid.UUID, _ string) (*models.User, error) {
	return nil, nil
}
func (s *stubUserRepo) ListByTenant(_ context.Context, _ uuid.UUID) ([]models.User, error) {
	return nil, nil
}
func (s *stubUserRepo) SoftDelete(_ context.Context, _, _ uuid.UUID) error { return nil }
func (s *stubUserRepo) Update(_ context.Context, _ *models.User) error     { return nil }
func (s *stubUserRepo) WithTx(_ db.DBTX) repository.UserRepository         { return s }

type stubTierRepo struct{}

func (s *stubTierRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Tier, error) {
	return &models.Tier{ID: uuid.New(), Name: "Standard"}, nil
}
func (s *stubTierRepo) GetByName(_ context.Context, _ string) (*models.Tier, error) {
	return &models.Tier{ID: uuid.New(), Name: "Standard"}, nil
}
func (s *stubTierRepo) List(_ context.Context) ([]models.Tier, error) { return nil, nil }

func newTenantHandler() *handler.TenantHandler {
	svc := service.NewTenantService(&stubTenantRepo{}, &stubTierRepo{}, &stubWalletRepo{}, &stubUserRepo{})
	return handler.NewTenantHandler(svc, slog.Default())
}

func TestOnboard_MissingFields(t *testing.T) {
	h := newTenantHandler()
	body, _ := json.Marshal(map[string]string{"name": "Acme"})
	req := httptest.NewRequest(http.MethodPost, "/tenants/onboard", bytes.NewReader(body))
	req = req.WithContext(middleware.WithUserID(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()
	h.Onboard(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestOnboard_DuplicateSlug_Returns409(t *testing.T) {
	createErr := &pgconn.PgError{Code: "23505", ConstraintName: "tenants_slug_key"}
	svc := service.NewTenantService(&stubTenantRepo{createErr: createErr}, &stubTierRepo{}, &stubWalletRepo{}, &stubUserRepo{})
	h := handler.NewTenantHandler(svc, slog.Default())

	body, _ := json.Marshal(map[string]any{
		"name":        "Acme",
		"admin_email": "admin@acme.com",
	})
	req := httptest.NewRequest(http.MethodPost, "/tenants/onboard", bytes.NewReader(body))
	req = req.WithContext(middleware.WithUserID(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()
	h.Onboard(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestOnboard_Valid(t *testing.T) {
	h := newTenantHandler()
	body, _ := json.Marshal(map[string]any{
		"name":        "Acme",
		"admin_email": "admin@acme.com",
	})
	req := httptest.NewRequest(http.MethodPost, "/tenants/onboard", bytes.NewReader(body))
	req = req.WithContext(middleware.WithUserID(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()
	h.Onboard(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func newTenantHandlerWithTenant(tenant *models.Tenant) *handler.TenantHandler {
	svc := service.NewTenantService(&stubTenantRepo{tenant: tenant}, &stubTierRepo{}, &stubWalletRepo{}, &stubUserRepo{})
	return handler.NewTenantHandler(svc, slog.Default())
}

func TestUpdateProfile_Valid(t *testing.T) {
	tenant := &models.Tenant{ID: uuid.New(), Name: "Acme", Status: models.TenantStatusActive}
	h := newTenantHandlerWithTenant(tenant)

	email := "contact@acme.com"
	body, _ := json.Marshal(map[string]any{
		"name":          "Acme Corp",
		"contact_email": email,
	})
	req := httptest.NewRequest(http.MethodPut, "/tenants/me", bytes.NewReader(body))
	req = req.WithContext(injectTenant(req.Context(), tenant))
	rec := httptest.NewRecorder()

	h.UpdateProfile(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateProfile_MissingName(t *testing.T) {
	tenant := &models.Tenant{ID: uuid.New(), Name: "Acme", Status: models.TenantStatusActive}
	h := newTenantHandlerWithTenant(tenant)

	body, _ := json.Marshal(map[string]any{"contact_email": "a@b.com"})
	req := httptest.NewRequest(http.MethodPut, "/tenants/me", bytes.NewReader(body))
	req = req.WithContext(injectTenant(req.Context(), tenant))
	rec := httptest.NewRecorder()

	h.UpdateProfile(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateProfile_InvalidEmail(t *testing.T) {
	tenant := &models.Tenant{ID: uuid.New(), Name: "Acme", Status: models.TenantStatusActive}
	h := newTenantHandlerWithTenant(tenant)

	body, _ := json.Marshal(map[string]any{
		"name":          "Acme",
		"contact_email": "not-an-email",
	})
	req := httptest.NewRequest(http.MethodPut, "/tenants/me", bytes.NewReader(body))
	req = req.WithContext(injectTenant(req.Context(), tenant))
	rec := httptest.NewRecorder()

	h.UpdateProfile(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateStorefront_Valid(t *testing.T) {
	tenant := &models.Tenant{ID: uuid.New(), Name: "Acme", Slug: "acme", Status: models.TenantStatusActive}
	h := newTenantHandlerWithTenant(tenant)

	body, _ := json.Marshal(map[string]any{
		"slug":                 "acme-store",
		"storefront_published": true,
	})
	req := httptest.NewRequest(http.MethodPut, "/tenants/me/storefront", bytes.NewReader(body))
	req = req.WithContext(middleware.WithUserRole(injectTenant(req.Context(), tenant), models.UserRoleAdmin))
	rec := httptest.NewRecorder()

	h.UpdateStorefront(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateStorefront_RejectsNonAdmin(t *testing.T) {
	tenant := &models.Tenant{ID: uuid.New(), Name: "Acme", Slug: "acme", Status: models.TenantStatusActive}
	h := newTenantHandlerWithTenant(tenant)

	body, _ := json.Marshal(map[string]any{
		"slug":                 "acme-store",
		"storefront_published": true,
	})
	req := httptest.NewRequest(http.MethodPut, "/tenants/me/storefront", bytes.NewReader(body))
	req = req.WithContext(middleware.WithUserRole(injectTenant(req.Context(), tenant), models.UserRoleStaff))
	rec := httptest.NewRecorder()

	h.UpdateStorefront(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateStorefront_RejectsReservedSlug(t *testing.T) {
	tenant := &models.Tenant{ID: uuid.New(), Name: "Acme", Slug: "acme", Status: models.TenantStatusActive}
	h := newTenantHandlerWithTenant(tenant)

	body, _ := json.Marshal(map[string]any{
		"slug":                 "app",
		"storefront_published": true,
	})
	req := httptest.NewRequest(http.MethodPut, "/tenants/me/storefront", bytes.NewReader(body))
	req = req.WithContext(middleware.WithUserRole(injectTenant(req.Context(), tenant), models.UserRoleAdmin))
	rec := httptest.NewRecorder()

	h.UpdateStorefront(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}
func injectTenant(ctx context.Context, t *models.Tenant) context.Context {
	return middleware.WithTenant(ctx, t)
}

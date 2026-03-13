package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"storefront/backend/internal/handler"
	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

// ── minimal mock repos for TenantService ─────────────────────────────────────

type stubTenantRepo struct {
	tenant    *models.Tenant
	createErr error
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
	s.tenant = t
	return nil
}
func (s *stubTenantRepo) SoftDelete(_ context.Context, _ uuid.UUID) error { return nil }

type stubWalletRepo struct{}

func (s *stubWalletRepo) Create(_ context.Context, w *models.Wallet) error {
	w.ID = uuid.New()
	return nil
}
func (s *stubWalletRepo) GetByTenantID(_ context.Context, _ uuid.UUID) (*models.Wallet, error) {
	return &models.Wallet{}, nil
}
func (s *stubWalletRepo) UpdateBalances(_ context.Context, _ *models.Wallet) error { return nil }

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

func TestGetMe(t *testing.T) {
	h := newTenantHandler()
	tenant := &models.Tenant{ID: uuid.New(), Name: "Acme", Status: models.TenantStatusActive}

	req := httptest.NewRequest(http.MethodGet, "/tenants/me", nil)
	req = req.WithContext(injectTenant(req.Context(), tenant))
	rec := httptest.NewRecorder()

	h.GetMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got models.Tenant
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if got.ID != tenant.ID {
		t.Fatalf("wrong tenant in response")
	}
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
	// Duplicate slug must return 409 Conflict, not 500 — the service detects the DB constraint name.
	createErr := errors.New(`ERROR: duplicate key value violates unique constraint "tenants_slug_key" (SQLSTATE 23505)`)
	svc := service.NewTenantService(&stubTenantRepo{createErr: createErr}, &stubTierRepo{}, &stubWalletRepo{}, &stubUserRepo{})
	h := handler.NewTenantHandler(svc, slog.Default())

	body, _ := json.Marshal(map[string]any{
		"name":        "Acme",
		"slug":        "acme",
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
		"slug":        "acme",
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

// injectTenant places the tenant into ctx using the middleware helper.
func injectTenant(ctx context.Context, t *models.Tenant) context.Context {
	return middleware.WithTenant(ctx, t)
}

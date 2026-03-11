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

	"storefront/backend/internal/handler"
	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

// ── minimal mock repos for TenantService ─────────────────────────────────────

type stubTenantRepo struct{ tenant *models.Tenant }

func (s *stubTenantRepo) Create(_ context.Context, t *models.Tenant) error {
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
func (s *stubUserRepo) SoftDelete(_ context.Context, _ uuid.UUID) error { return nil }

func newTenantHandler() *handler.TenantHandler {
	svc := service.NewTenantService(&stubTenantRepo{}, &stubWalletRepo{}, &stubUserRepo{})
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
	rec := httptest.NewRecorder()
	h.Onboard(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestOnboard_Valid(t *testing.T) {
	h := newTenantHandler()
	body, _ := json.Marshal(map[string]any{
		"name":          "Acme",
		"slug":          "acme",
		"tier_id":       uuid.New(),
		"admin_user_id": uuid.New(),
		"admin_email":   "admin@acme.com",
	})
	req := httptest.NewRequest(http.MethodPost, "/tenants/onboard", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Onboard(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

// injectTenant places the tenant into ctx using the middleware helper.
func injectTenant(ctx context.Context, t *models.Tenant) context.Context {
	return middleware.WithTenant(ctx, t)
}

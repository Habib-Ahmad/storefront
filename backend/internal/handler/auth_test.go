package handler_test

import (
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
	"storefront/backend/internal/repository"
)

type authUserRepo struct {
	user *models.User
	err  error
}

func (m *authUserRepo) Create(_ context.Context, _ *models.User) error { return nil }
func (m *authUserRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.User, error) {
	return m.user, m.err
}
func (m *authUserRepo) GetByEmail(_ context.Context, _ uuid.UUID, _ string) (*models.User, error) {
	return nil, nil
}
func (m *authUserRepo) ListByTenant(_ context.Context, _ uuid.UUID) ([]models.User, error) {
	return nil, nil
}
func (m *authUserRepo) Update(_ context.Context, _ *models.User) error     { return nil }
func (m *authUserRepo) SoftDelete(_ context.Context, _, _ uuid.UUID) error { return nil }

var _ repository.UserRepository = (*authUserRepo)(nil)

type authTenantRepo struct {
	tenant *models.Tenant
	err    error
}

func (m *authTenantRepo) Create(_ context.Context, _ *models.Tenant) error { return nil }
func (m *authTenantRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Tenant, error) {
	return m.tenant, m.err
}
func (m *authTenantRepo) GetBySlug(_ context.Context, _ string) (*models.Tenant, error) {
	return nil, nil
}
func (m *authTenantRepo) Update(_ context.Context, _ *models.Tenant) error { return nil }
func (m *authTenantRepo) SoftDelete(_ context.Context, _ uuid.UUID) error  { return nil }

var _ repository.TenantRepository = (*authTenantRepo)(nil)

func TestAuthMe_NotOnboarded(t *testing.T) {
	users := &authUserRepo{err: errors.New("not found")}
	tenants := &authTenantRepo{}
	h := handler.NewAuthHandler(users, tenants, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body["onboarded"] != false {
		t.Fatal("expected onboarded=false")
	}
}

func TestAuthMe_Onboarded(t *testing.T) {
	tenantID := uuid.New()
	users := &authUserRepo{user: &models.User{
		ID:       uuid.New(),
		TenantID: tenantID,
		Email:    "admin@acme.com",
		Role:     models.UserRoleAdmin,
	}}
	tenants := &authTenantRepo{tenant: &models.Tenant{
		ID:     tenantID,
		Name:   "Acme",
		Status: models.TenantStatusActive,
	}}
	h := handler.NewAuthHandler(users, tenants, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body struct {
		Onboarded bool            `json:"onboarded"`
		Role      models.UserRole `json:"role"`
		Tenant    models.Tenant   `json:"tenant"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if !body.Onboarded {
		t.Fatal("expected onboarded=true")
	}
	if body.Role != models.UserRoleAdmin {
		t.Fatalf("expected role admin, got %s", body.Role)
	}
	if body.Tenant.ID != tenantID {
		t.Fatalf("expected tenant %s, got %s", tenantID, body.Tenant.ID)
	}
}

func TestAuthMe_UserExistsTenantMissing(t *testing.T) {
	users := &authUserRepo{user: &models.User{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Role:     models.UserRoleAdmin,
	}}
	tenants := &authTenantRepo{err: errors.New("not found")}
	h := handler.NewAuthHandler(users, tenants, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(middleware.WithUserID(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body["onboarded"] != false {
		t.Fatal("expected onboarded=false when tenant is missing")
	}
}

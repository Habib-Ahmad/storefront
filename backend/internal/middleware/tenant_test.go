package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"storefront/backend/internal/db"
	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

type mockUserRepo struct {
	user *models.User
	err  error
}

func (m *mockUserRepo) Create(_ context.Context, _ *models.User) error { return nil }
func (m *mockUserRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.User, error) {
	return m.user, m.err
}
func (m *mockUserRepo) GetByEmail(_ context.Context, _ uuid.UUID, _ string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) ListByTenant(_ context.Context, _ uuid.UUID) ([]models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Update(_ context.Context, _ *models.User) error     { return nil }
func (m *mockUserRepo) SoftDelete(_ context.Context, _, _ uuid.UUID) error { return nil }
func (m *mockUserRepo) WithTx(_ db.DBTX) repository.UserRepository         { return m }

type mockTenantRepo struct {
	tenant *models.Tenant
	err    error
}

func (m *mockTenantRepo) Create(_ context.Context, _ *models.Tenant) error { return nil }
func (m *mockTenantRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Tenant, error) {
	return m.tenant, m.err
}
func (m *mockTenantRepo) GetBySlug(_ context.Context, _ string) (*models.Tenant, error) {
	return nil, nil
}
func (m *mockTenantRepo) Update(_ context.Context, _ *models.Tenant) error { return nil }
func (m *mockTenantRepo) SoftDelete(_ context.Context, _ uuid.UUID) error  { return nil }
func (m *mockTenantRepo) WithTx(_ db.DBTX) repository.TenantRepository     { return m }

func reqWithUserID(userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	return req.WithContext(middleware.WithUserID(req.Context(), userID))
}

func TestResolveTenant_Active(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()

	users := &mockUserRepo{user: &models.User{ID: userID, TenantID: tenantID, Role: models.UserRoleAdmin}}
	tenants := &mockTenantRepo{tenant: &models.Tenant{ID: tenantID, Status: models.TenantStatusActive}}

	var gotTenant *models.Tenant
	var gotRole models.UserRole
	mw := middleware.ResolveTenant(users, tenants)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTenant = middleware.TenantFromCtx(r.Context())
		gotRole = middleware.UserRoleFromCtx(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, reqWithUserID(userID))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotTenant == nil || gotTenant.ID != tenantID {
		t.Fatal("tenant not injected into context")
	}
	if gotRole != models.UserRoleAdmin {
		t.Fatalf("expected role admin in context, got %s", gotRole)
	}
}

func TestResolveTenant_OverridesTokenRoleWithDatabaseRole(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()

	users := &mockUserRepo{user: &models.User{ID: userID, TenantID: tenantID, Role: models.UserRoleAdmin}}
	tenants := &mockTenantRepo{tenant: &models.Tenant{ID: tenantID, Status: models.TenantStatusActive}}

	var gotRole models.UserRole
	mw := middleware.ResolveTenant(users, tenants)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRole = middleware.UserRoleFromCtx(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := reqWithUserID(userID)
	req = req.WithContext(middleware.WithUserRole(req.Context(), models.UserRoleStaff))
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotRole != models.UserRoleAdmin {
		t.Fatalf("expected database role admin in context, got %s", gotRole)
	}
}

func TestResolveTenant_Suspended(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()

	users := &mockUserRepo{user: &models.User{ID: userID, TenantID: tenantID}}
	tenants := &mockTenantRepo{tenant: &models.Tenant{ID: tenantID, Status: models.TenantStatusSuspended}}

	mw := middleware.ResolveTenant(users, tenants)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, reqWithUserID(userID))

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestResolveTenant_UserNotFound(t *testing.T) {
	users := &mockUserRepo{err: errors.New("not found")}
	tenants := &mockTenantRepo{}

	mw := middleware.ResolveTenant(users, tenants)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, reqWithUserID(uuid.New()))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestResolveTenant_TenantNotFound(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()

	users := &mockUserRepo{user: &models.User{ID: userID, TenantID: tenantID}}
	tenants := &mockTenantRepo{err: errors.New("not found")}

	mw := middleware.ResolveTenant(users, tenants)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, reqWithUserID(userID))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"storefront/backend/internal/authz"
	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
)

func TestRequirePermission_AllowsGrantedRole(t *testing.T) {
	handler := middleware.RequirePermission(authz.PermissionProductsManage)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPost, "/products", nil)
	req = req.WithContext(middleware.WithUserRole(req.Context(), models.UserRoleStaff))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestRequirePermission_RejectsForbiddenRole(t *testing.T) {
	handler := middleware.RequirePermission(authz.PermissionStorefrontManage)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodPut, "/tenants/me/storefront", nil)
	req = req.WithContext(middleware.WithUserRole(req.Context(), models.UserRoleStaff))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

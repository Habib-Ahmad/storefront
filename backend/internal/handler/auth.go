package handler

import (
	"log/slog"
	"net/http"

	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
)

type AuthHandler struct {
	users   repository.UserRepository
	tenants repository.TenantRepository
	log     *slog.Logger
}

func NewAuthHandler(users repository.UserRepository, tenants repository.TenantRepository, log *slog.Logger) *AuthHandler {
	return &AuthHandler{users: users, tenants: tenants, log: log}
}

// GET /auth/me — tells the frontend whether the authenticated user has completed onboarding.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())

	user, err := h.users.GetByID(r.Context(), userID)
	if err != nil || user == nil {
		respond(w, http.StatusOK, map[string]any{"onboarded": false})
		return
	}

	tenant, err := h.tenants.GetByID(r.Context(), user.TenantID)
	if err != nil || tenant == nil {
		respond(w, http.StatusOK, map[string]any{"onboarded": false})
		return
	}

	respond(w, http.StatusOK, struct {
		Onboarded bool            `json:"onboarded"`
		Tenant    *models.Tenant  `json:"tenant"`
		Role      models.UserRole `json:"role"`
	}{
		Onboarded: true,
		Tenant:    tenant,
		Role:      user.Role,
	})
}

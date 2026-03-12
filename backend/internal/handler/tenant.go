package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

type TenantHandler struct {
	svc *service.TenantService
	log *slog.Logger
}

func NewTenantHandler(svc *service.TenantService, log *slog.Logger) *TenantHandler {
	return &TenantHandler{svc: svc, log: log}
}

// GET /tenants/me
func (h *TenantHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	respond(w, http.StatusOK, tenant)
}

// POST /tenants/onboard
func (h *TenantHandler) Onboard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromCtx(r.Context())
	if userID == uuid.Nil {
		respondErr(w, http.StatusUnauthorized, "missing user identity")
		return
	}

	var req struct {
		Name       string    `json:"name"         validate:"required"`
		Slug       string    `json:"slug"         validate:"required"`
		TierID     uuid.UUID `json:"tier_id"`
		AdminEmail string    `json:"admin_email"  validate:"required,email"`
	}
	if !decodeValid(w, r, &req) {
		return
	}

	tenant, err := h.svc.Onboard(r.Context(), req.Name, req.Slug, req.TierID, userID, req.AdminEmail)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlugTaken):
			respondErr(w, http.StatusConflict, "slug already in use")
		case errors.Is(err, service.ErrUserExists):
			respondErr(w, http.StatusConflict, "user already belongs to a tenant")
		default:
			serverErr(w, h.log, r, err)
		}
		return
	}
	respond(w, http.StatusCreated, tenant)
}

// PUT /tenants/me
func (h *TenantHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	var req struct {
		Name string `json:"name" validate:"required"`
	}
	if !decodeValid(w, r, &req) {
		return
	}
	if err := h.svc.UpdateProfile(r.Context(), tenant.ID, req.Name); err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PUT /tenants/me/modules
func (h *TenantHandler) SetModules(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	var mods models.ActiveModules
	if !decodeValid(w, r, &mods) {
		return
	}
	if err := h.svc.SetModules(r.Context(), tenant.ID, mods); err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

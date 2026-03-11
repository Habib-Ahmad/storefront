package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// POST /products
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	if err := service.RequireModule(tenant, true, false, false); err != nil {
		respondErr(w, http.StatusForbidden, "inventory module not enabled")
		return
	}

	var req struct {
		Name        string                  `json:"name"`
		Description *string                 `json:"description"`
		Category    *string                 `json:"category"`
		IsAvailable bool                    `json:"is_available"`
		Variants    []models.ProductVariant `json:"variants"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		respondErr(w, http.StatusBadRequest, "name is required")
		return
	}

	p := &models.Product{
		TenantID:    tenant.ID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		IsAvailable: req.IsAvailable,
	}
	out, err := h.svc.Create(r.Context(), p, req.Variants)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, "create failed")
		return
	}
	respond(w, http.StatusCreated, out)
}

// GET /products
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	if err := service.RequireModule(tenant, true, false, false); err != nil {
		respondErr(w, http.StatusForbidden, "inventory module not enabled")
		return
	}
	_ = tenant // repo call goes here in Block 6 wiring — handler shells are intentional
	respond(w, http.StatusOK, []any{})
}

// DELETE /products/{id}
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid product id")
		return
	}
	if err := h.svc.SoftDelete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			respondErr(w, http.StatusNotFound, "product not found")
			return
		}
		respondErr(w, http.StatusInternalServerError, "delete failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

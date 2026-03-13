package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

type ProductHandler struct {
	svc *service.ProductService
	log *slog.Logger
}

func NewProductHandler(svc *service.ProductService, log *slog.Logger) *ProductHandler {
	return &ProductHandler{svc: svc, log: log}
}

// POST /products
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	if err := service.RequireModule(tenant, true, false, false); err != nil {
		respondErr(w, http.StatusForbidden, "inventory module not enabled")
		return
	}

	var req struct {
		Name        string                  `json:"name"        validate:"required"`
		Description *string                 `json:"description"`
		Category    *string                 `json:"category"`
		IsAvailable bool                    `json:"is_available"`
		Variants    []models.ProductVariant `json:"variants"`
	}
	if !decodeValid(w, r, &req) {
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
		serverErr(w, h.log, r, err)
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
	products, err := h.svc.List(r.Context(), tenant.ID)
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	if products == nil {
		products = []models.Product{}
	}
	respond(w, http.StatusOK, products)
}

// GET /products/{id}
func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	if err := service.RequireModule(tenant, true, false, false); err != nil {
		respondErr(w, http.StatusForbidden, "inventory module not enabled")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid product id")
		return
	}
	p, variants, err := h.svc.GetByID(r.Context(), tenant.ID, id)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			respondErr(w, http.StatusNotFound, "product not found")
			return
		}
		serverErr(w, h.log, r, err)
		return
	}
	respond(w, http.StatusOK, map[string]any{"product": p, "variants": variants})
}

// PUT /products/{id}
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	if err := service.RequireModule(tenant, true, false, false); err != nil {
		respondErr(w, http.StatusForbidden, "inventory module not enabled")
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid product id")
		return
	}
	var req struct {
		Name        string  `json:"name"        validate:"required"`
		Description *string `json:"description"`
		Category    *string `json:"category"`
		IsAvailable bool    `json:"is_available"`
	}
	if !decodeValid(w, r, &req) {
		return
	}
	p := &models.Product{
		ID:          id,
		TenantID:    tenant.ID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		IsAvailable: req.IsAvailable,
	}
	if err := h.svc.Update(r.Context(), p); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			respondErr(w, http.StatusNotFound, "product not found")
			return
		}
		serverErr(w, h.log, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /products/{id}
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	if err := service.RequireModule(tenant, true, false, false); err != nil {
		respondErr(w, http.StatusForbidden, "inventory module not enabled")
		return
	}
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid product id")
		return
	}
	if err := h.svc.SoftDelete(r.Context(), tenant.ID, id); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			respondErr(w, http.StatusNotFound, "product not found")
			return
		}
		respondErr(w, http.StatusInternalServerError, "delete failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /products/{id}/variants
func (h *ProductHandler) CreateVariant(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	if err := service.RequireModule(tenant, true, false, false); err != nil {
		respondErr(w, http.StatusForbidden, "inventory module not enabled")
		return
	}
	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid product id")
		return
	}
	var req struct {
		SKU        string          `json:"sku"        validate:"required"`
		Attributes json.RawMessage `json:"attributes"`
		Price      decimal.Decimal `json:"price"      validate:"required"`
		StockQty   *int            `json:"stock_qty"`
	}
	if !decodeValid(w, r, &req) {
		return
	}
	v := &models.ProductVariant{
		ProductID:  productID,
		SKU:        req.SKU,
		Attributes: req.Attributes,
		Price:      req.Price,
		StockQty:   req.StockQty,
	}
	if err := h.svc.CreateVariant(r.Context(), tenant.ID, v); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			respondErr(w, http.StatusNotFound, "product not found")
			return
		}
		serverErr(w, h.log, r, err)
		return
	}
	respond(w, http.StatusCreated, v)
}

// GET /products/{id}/variants
func (h *ProductHandler) ListVariants(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	if err := service.RequireModule(tenant, true, false, false); err != nil {
		respondErr(w, http.StatusForbidden, "inventory module not enabled")
		return
	}
	productID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid product id")
		return
	}
	variants, err := h.svc.ListVariants(r.Context(), tenant.ID, productID)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			respondErr(w, http.StatusNotFound, "product not found")
			return
		}
		serverErr(w, h.log, r, err)
		return
	}
	if variants == nil {
		variants = []models.ProductVariant{}
	}
	respond(w, http.StatusOK, variants)
}

// PUT /products/{id}/variants/{variantId}
func (h *ProductHandler) UpdateVariant(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	if err := service.RequireModule(tenant, true, false, false); err != nil {
		respondErr(w, http.StatusForbidden, "inventory module not enabled")
		return
	}
	variantID, err := uuid.Parse(chi.URLParam(r, "variantId"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid variant id")
		return
	}
	var req struct {
		SKU        string          `json:"sku"        validate:"required"`
		Attributes json.RawMessage `json:"attributes"`
		Price      decimal.Decimal `json:"price"      validate:"required"`
		StockQty   *int            `json:"stock_qty"`
	}
	if !decodeValid(w, r, &req) {
		return
	}
	v := &models.ProductVariant{
		ID:         variantID,
		SKU:        req.SKU,
		Attributes: req.Attributes,
		Price:      req.Price,
		StockQty:   req.StockQty,
	}
	if err := h.svc.UpdateVariant(r.Context(), tenant.ID, v); err != nil {
		if errors.Is(err, service.ErrVariantNotFound) {
			respondErr(w, http.StatusNotFound, "variant not found")
			return
		}
		serverErr(w, h.log, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /products/{id}/variants/{variantId}
func (h *ProductHandler) DeleteVariant(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromCtx(r.Context())
	if err := service.RequireModule(tenant, true, false, false); err != nil {
		respondErr(w, http.StatusForbidden, "inventory module not enabled")
		return
	}
	variantID, err := uuid.Parse(chi.URLParam(r, "variantId"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid variant id")
		return
	}
	if err := h.svc.DeleteVariant(r.Context(), tenant.ID, variantID); err != nil {
		if errors.Is(err, service.ErrVariantNotFound) {
			respondErr(w, http.StatusNotFound, "variant not found")
			return
		}
		serverErr(w, h.log, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

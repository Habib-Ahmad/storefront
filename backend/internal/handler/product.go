package handler

import (
	"encoding/json"
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
	svc   *service.ProductService
	media *MediaHandler
	log   *slog.Logger
}

func NewProductHandler(svc *service.ProductService, media *MediaHandler, log *slog.Logger) *ProductHandler {
	return &ProductHandler{svc: svc, media: media, log: log}
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
		handleErr(w, h.log, r, err)
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
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)
	products, err := h.svc.List(r.Context(), tenant.ID, limit, offset)
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	for i := range products {
		variants, err := h.svc.ListVariants(r.Context(), tenant.ID, products[i].ID)
		if err != nil {
			serverErr(w, h.log, r, err)
			return
		}
		images, err := h.svc.GetImagesByProduct(r.Context(), tenant.ID, products[i].ID)
		if err != nil {
			serverErr(w, h.log, r, err)
			return
		}
		if variants == nil {
			variants = []models.ProductVariant{}
		}
		if images == nil {
			images = []models.ProductImage{}
		}
		products[i].Variants = variants
		products[i].Images = images
	}
	if products == nil {
		products = []models.Product{}
	}
	total, err := h.svc.CountByTenant(r.Context(), tenant.ID)
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	respondPage(w, products, total, limit, offset)
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
		handleErr(w, h.log, r, err)
		return
	}
	images, _ := h.svc.GetImagesByProduct(r.Context(), tenant.ID, id)
	if images == nil {
		images = []models.ProductImage{}
	}
	respond(w, http.StatusOK, map[string]any{"product": p, "variants": variants, "images": images})
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
		handleErr(w, h.log, r, err)
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
		handleErr(w, h.log, r, err)
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
		SKU        string           `json:"sku"        validate:"required"`
		Attributes json.RawMessage  `json:"attributes"`
		Price      decimal.Decimal  `json:"price"      validate:"required"`
		CostPrice  *decimal.Decimal `json:"cost_price"`
		StockQty   *int             `json:"stock_qty"`
	}
	if !decodeValid(w, r, &req) {
		return
	}
	v := &models.ProductVariant{
		ProductID:  productID,
		SKU:        req.SKU,
		Attributes: req.Attributes,
		Price:      req.Price,
		CostPrice:  req.CostPrice,
		StockQty:   req.StockQty,
	}
	if err := h.svc.CreateVariant(r.Context(), tenant.ID, v); err != nil {
		handleErr(w, h.log, r, err)
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
		handleErr(w, h.log, r, err)
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
		SKU        string           `json:"sku"        validate:"required"`
		Attributes json.RawMessage  `json:"attributes"`
		Price      decimal.Decimal  `json:"price"      validate:"required"`
		CostPrice  *decimal.Decimal `json:"cost_price"`
		StockQty   *int             `json:"stock_qty"`
	}
	if !decodeValid(w, r, &req) {
		return
	}
	v := &models.ProductVariant{
		ID:         variantID,
		SKU:        req.SKU,
		Attributes: req.Attributes,
		Price:      req.Price,
		CostPrice:  req.CostPrice,
		StockQty:   req.StockQty,
	}
	if err := h.svc.UpdateVariant(r.Context(), tenant.ID, v); err != nil {
		handleErr(w, h.log, r, err)
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
		handleErr(w, h.log, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /products/{id}/images
func (h *ProductHandler) AddImage(w http.ResponseWriter, r *http.Request) {
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
		URL       string `json:"url"       validate:"required,url"`
		SortOrder int    `json:"sort_order"`
		IsPrimary bool   `json:"is_primary"`
	}
	if !decodeValid(w, r, &req) {
		return
	}
	img := &models.ProductImage{
		ProductID: productID,
		URL:       req.URL,
		SortOrder: req.SortOrder,
		IsPrimary: req.IsPrimary,
	}
	if err := h.svc.AddImage(r.Context(), tenant.ID, img); err != nil {
		handleErr(w, h.log, r, err)
		return
	}
	respond(w, http.StatusCreated, img)
}

// GET /products/{id}/images
func (h *ProductHandler) ListImages(w http.ResponseWriter, r *http.Request) {
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
	images, err := h.svc.GetImagesByProduct(r.Context(), tenant.ID, productID)
	if err != nil {
		handleErr(w, h.log, r, err)
		return
	}
	if images == nil {
		images = []models.ProductImage{}
	}
	respond(w, http.StatusOK, images)
}

// PUT /products/{id}/images/{imageId}
func (h *ProductHandler) UpdateImage(w http.ResponseWriter, r *http.Request) {
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
	imageID, err := uuid.Parse(chi.URLParam(r, "imageId"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid image id")
		return
	}
	var req struct {
		URL       string `json:"url"       validate:"required,url"`
		SortOrder int    `json:"sort_order"`
		IsPrimary bool   `json:"is_primary"`
	}
	if !decodeValid(w, r, &req) {
		return
	}
	img := &models.ProductImage{
		ID:        imageID,
		ProductID: productID,
		URL:       req.URL,
		SortOrder: req.SortOrder,
		IsPrimary: req.IsPrimary,
	}
	if err := h.svc.UpdateImage(r.Context(), tenant.ID, img); err != nil {
		handleErr(w, h.log, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /products/{id}/images/{imageId}
func (h *ProductHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
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
	imageID, err := uuid.Parse(chi.URLParam(r, "imageId"))
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid image id")
		return
	}
	img, err := h.svc.GetImage(r.Context(), tenant.ID, productID, imageID)
	if err != nil {
		handleErr(w, h.log, r, err)
		return
	}
	if err := h.media.DeleteObjectByURL(r.Context(), img.URL); err != nil {
		serverErr(w, h.log, r, err)
		return
	}
	if err := h.svc.DeleteImage(r.Context(), tenant.ID, productID, imageID); err != nil {
		handleErr(w, h.log, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

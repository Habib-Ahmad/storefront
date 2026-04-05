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

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/handler"
	"storefront/backend/internal/middleware"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
	"storefront/backend/internal/service"
)

type stubProductRepo struct {
	product        *models.Product
	products       []models.Product
	publicProducts []models.PublicStorefrontProduct
	variant        *models.ProductVariant
	variants       []models.ProductVariant
	createErr      error
	getErr         error
	addImageErr    error
	updateImageErr error
}

func (s *stubProductRepo) Create(_ context.Context, p *models.Product) error {
	if s.createErr != nil {
		return s.createErr
	}
	p.ID = uuid.New()
	s.product = p
	return nil
}
func (s *stubProductRepo) GetByID(_ context.Context, _, id uuid.UUID) (*models.Product, error) {
	if s.getErr != nil {
		return nil, s.getErr
	}
	if s.product != nil {
		return s.product, nil
	}
	return &models.Product{ID: id, IsAvailable: true}, nil
}
func (s *stubProductRepo) ListByTenant(_ context.Context, _ uuid.UUID, _, _ int) ([]models.Product, error) {
	return s.products, nil
}
func (s *stubProductRepo) ListPublicByTenant(_ context.Context, _ uuid.UUID) ([]models.PublicStorefrontProduct, error) {
	return s.publicProducts, nil
}
func (s *stubProductRepo) CountByTenant(_ context.Context, _ uuid.UUID) (int, error) {
	return len(s.products), nil
}
func (s *stubProductRepo) Update(_ context.Context, _ *models.Product) error  { return nil }
func (s *stubProductRepo) SoftDelete(_ context.Context, _, _ uuid.UUID) error { return nil }
func (s *stubProductRepo) CreateVariant(_ context.Context, v *models.ProductVariant) error {
	v.ID = uuid.New()
	return nil
}
func (s *stubProductRepo) GetVariantByID(_ context.Context, _ uuid.UUID) (*models.ProductVariant, error) {
	if s.variant == nil {
		return nil, errNotFound
	}
	return s.variant, nil
}
func (s *stubProductRepo) ListVariants(_ context.Context, _ uuid.UUID) ([]models.ProductVariant, error) {
	return s.variants, nil
}
func (s *stubProductRepo) UpdateVariant(_ context.Context, _ *models.ProductVariant) error {
	return nil
}
func (s *stubProductRepo) SoftDeleteVariant(_ context.Context, _ uuid.UUID) error     { return nil }
func (s *stubProductRepo) DecrementStock(_ context.Context, _ uuid.UUID, _ int) error { return nil }
func (s *stubProductRepo) RestoreStock(_ context.Context, _ uuid.UUID, _ int) error   { return nil }
func (s *stubProductRepo) AddImage(_ context.Context, img *models.ProductImage) error {
	if s.addImageErr != nil {
		return s.addImageErr
	}
	img.ID = uuid.New()
	return nil
}
func (s *stubProductRepo) ListImagesByProduct(_ context.Context, _ uuid.UUID) ([]models.ProductImage, error) {
	return nil, nil
}
func (s *stubProductRepo) DeleteImage(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubProductRepo) UpdateImage(_ context.Context, _ *models.ProductImage) error {
	return s.updateImageErr
}

var _ repository.ProductRepository = (*stubProductRepo)(nil)

var errNotFound = errors.New("not found")

func activeTenant() *models.Tenant {
	return &models.Tenant{
		ID:            uuid.New(),
		Name:          "Test Store",
		Status:        models.TenantStatusActive,
		ActiveModules: models.ActiveModules{Inventory: true, Payments: true},
	}
}

func noInventoryTenant() *models.Tenant {
	return &models.Tenant{
		ID:            uuid.New(),
		Name:          "No Inv",
		Status:        models.TenantStatusActive,
		ActiveModules: models.ActiveModules{Inventory: false},
	}
}

func newProductHandler(repo *stubProductRepo) *handler.ProductHandler {
	svc := service.NewProductService(repo)
	return handler.NewProductHandler(svc, slog.Default())
}

func withChiParam(req *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func withChiParams(req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestCreateProduct_Valid(t *testing.T) {
	h := newProductHandler(&stubProductRepo{})
	body, _ := json.Marshal(map[string]any{
		"name":         "Classic Tee",
		"description":  "100% cotton",
		"category":     "Fashion",
		"is_available": true,
	})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateProduct_MissingName(t *testing.T) {
	h := newProductHandler(&stubProductRepo{})
	body, _ := json.Marshal(map[string]any{"is_available": true})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateProduct_ModuleDisabled(t *testing.T) {
	h := newProductHandler(&stubProductRepo{})
	body, _ := json.Marshal(map[string]any{"name": "T-Shirt", "is_available": true})
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), noInventoryTenant()))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestListProducts_OK(t *testing.T) {
	repo := &stubProductRepo{products: []models.Product{
		{ID: uuid.New(), Name: "A"},
		{ID: uuid.New(), Name: "B"},
	}}
	h := newProductHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got struct {
		Data []models.Product `json:"data"`
	}
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 products, got %d", len(got.Data))
	}
}

func TestListProducts_Empty(t *testing.T) {
	h := newProductHandler(&stubProductRepo{})
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	rec := httptest.NewRecorder()
	h.List(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGetProduct_OK(t *testing.T) {
	productID := uuid.New()
	repo := &stubProductRepo{
		product:  &models.Product{ID: productID, Name: "Tee"},
		variants: []models.ProductVariant{{ID: uuid.New(), ProductID: productID, SKU: "TEE-M"}},
	}
	h := newProductHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/products/"+productID.String(), nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", productID.String())
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got map[string]json.RawMessage
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if _, ok := got["product"]; !ok {
		t.Fatal("response missing 'product' key")
	}
	if _, ok := got["variants"]; !ok {
		t.Fatal("response missing 'variants' key")
	}
}

func TestGetProduct_NotFound(t *testing.T) {
	repo := &stubProductRepo{getErr: errNotFound}
	h := newProductHandler(repo)
	id := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/products/"+id.String(), nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", id.String())
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetProduct_InvalidID(t *testing.T) {
	h := newProductHandler(&stubProductRepo{})
	req := httptest.NewRequest(http.MethodGet, "/products/not-a-uuid", nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", "not-a-uuid")
	rec := httptest.NewRecorder()
	h.Get(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestUpdateProduct_OK(t *testing.T) {
	productID := uuid.New()
	repo := &stubProductRepo{product: &models.Product{ID: productID, Name: "Old"}}
	h := newProductHandler(repo)
	body, _ := json.Marshal(map[string]any{
		"name":         "Updated Tee",
		"is_available": true,
	})
	req := httptest.NewRequest(http.MethodPut, "/products/"+productID.String(), bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", productID.String())
	rec := httptest.NewRecorder()
	h.Update(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateProduct_NotFound(t *testing.T) {
	repo := &stubProductRepo{getErr: errNotFound}
	h := newProductHandler(repo)
	id := uuid.New()
	body, _ := json.Marshal(map[string]any{"name": "X", "is_available": true})
	req := httptest.NewRequest(http.MethodPut, "/products/"+id.String(), bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", id.String())
	rec := httptest.NewRecorder()
	h.Update(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteProduct_OK(t *testing.T) {
	productID := uuid.New()
	repo := &stubProductRepo{product: &models.Product{ID: productID}}
	h := newProductHandler(repo)
	req := httptest.NewRequest(http.MethodDelete, "/products/"+productID.String(), nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", productID.String())
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteProduct_ModuleDisabled(t *testing.T) {
	h := newProductHandler(&stubProductRepo{})
	req := httptest.NewRequest(http.MethodDelete, "/products/"+uuid.New().String(), nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), noInventoryTenant()))
	req = withChiParam(req, "id", uuid.New().String())
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestDeleteProduct_InvalidID(t *testing.T) {
	h := newProductHandler(&stubProductRepo{})
	req := httptest.NewRequest(http.MethodDelete, "/products/bad", nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", "bad")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateVariant_OK(t *testing.T) {
	productID := uuid.New()
	repo := &stubProductRepo{product: &models.Product{ID: productID}}
	h := newProductHandler(repo)
	body, _ := json.Marshal(map[string]any{
		"sku":        "TEE-RED-M",
		"attributes": map[string]string{"color": "red", "size": "M"},
		"price":      5500,
		"stock_qty":  30,
	})
	req := httptest.NewRequest(http.MethodPost, "/products/"+productID.String()+"/variants", bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", productID.String())
	rec := httptest.NewRecorder()
	h.CreateVariant(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateVariant_ProductNotFound(t *testing.T) {
	repo := &stubProductRepo{getErr: errNotFound}
	h := newProductHandler(repo)
	body, _ := json.Marshal(map[string]any{
		"sku":   "TEE-M",
		"price": 1000,
	})
	req := httptest.NewRequest(http.MethodPost, "/products/"+uuid.New().String()+"/variants", bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", uuid.New().String())
	rec := httptest.NewRecorder()
	h.CreateVariant(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateVariant_MissingSKU(t *testing.T) {
	h := newProductHandler(&stubProductRepo{})
	body, _ := json.Marshal(map[string]any{"price": 1000})
	req := httptest.NewRequest(http.MethodPost, "/products/"+uuid.New().String()+"/variants", bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", uuid.New().String())
	rec := httptest.NewRecorder()
	h.CreateVariant(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestCreateVariant_ModuleDisabled(t *testing.T) {
	h := newProductHandler(&stubProductRepo{})
	body, _ := json.Marshal(map[string]any{"sku": "X", "price": 1000})
	req := httptest.NewRequest(http.MethodPost, "/products/"+uuid.New().String()+"/variants", bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), noInventoryTenant()))
	req = withChiParam(req, "id", uuid.New().String())
	rec := httptest.NewRecorder()
	h.CreateVariant(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestListVariants_OK(t *testing.T) {
	productID := uuid.New()
	repo := &stubProductRepo{
		product: &models.Product{ID: productID},
		variants: []models.ProductVariant{
			{ID: uuid.New(), ProductID: productID, SKU: "A"},
			{ID: uuid.New(), ProductID: productID, SKU: "B"},
		},
	}
	h := newProductHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/products/"+productID.String()+"/variants", nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", productID.String())
	rec := httptest.NewRecorder()
	h.ListVariants(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got []models.ProductVariant
	_ = json.NewDecoder(rec.Body).Decode(&got)
	if len(got) != 2 {
		t.Fatalf("expected 2 variants, got %d", len(got))
	}
}

func TestListVariants_ProductNotFound(t *testing.T) {
	repo := &stubProductRepo{getErr: errNotFound}
	h := newProductHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/products/"+uuid.New().String()+"/variants", nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", uuid.New().String())
	rec := httptest.NewRecorder()
	h.ListVariants(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUpdateVariant_OK(t *testing.T) {
	productID := uuid.New()
	variantID := uuid.New()
	repo := &stubProductRepo{
		product: &models.Product{ID: productID},
		variant: &models.ProductVariant{ID: variantID, ProductID: productID, SKU: "OLD", Price: decimal.NewFromInt(100)},
	}
	h := newProductHandler(repo)
	body, _ := json.Marshal(map[string]any{
		"sku":   "NEW-SKU",
		"price": 6000,
	})
	req := httptest.NewRequest(http.MethodPut, "/products/"+productID.String()+"/variants/"+variantID.String(), bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParams(req, map[string]string{"id": productID.String(), "variantId": variantID.String()})
	rec := httptest.NewRecorder()
	h.UpdateVariant(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateVariant_NotFound(t *testing.T) {
	repo := &stubProductRepo{variant: nil}
	h := newProductHandler(repo)
	body, _ := json.Marshal(map[string]any{"sku": "X", "price": 1000})
	req := httptest.NewRequest(http.MethodPut, "/products/"+uuid.New().String()+"/variants/"+uuid.New().String(), bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParams(req, map[string]string{"id": uuid.New().String(), "variantId": uuid.New().String()})
	rec := httptest.NewRecorder()
	h.UpdateVariant(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUpdateVariant_InvalidID(t *testing.T) {
	h := newProductHandler(&stubProductRepo{})
	body, _ := json.Marshal(map[string]any{"sku": "X", "price": 1000})
	req := httptest.NewRequest(http.MethodPut, "/products/abc/variants/bad", bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParams(req, map[string]string{"id": "abc", "variantId": "bad"})
	rec := httptest.NewRecorder()
	h.UpdateVariant(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteVariant_OK(t *testing.T) {
	productID := uuid.New()
	variantID := uuid.New()
	repo := &stubProductRepo{
		product: &models.Product{ID: productID},
		variant: &models.ProductVariant{ID: variantID, ProductID: productID},
	}
	h := newProductHandler(repo)
	req := httptest.NewRequest(http.MethodDelete, "/products/"+productID.String()+"/variants/"+variantID.String(), nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParams(req, map[string]string{"id": productID.String(), "variantId": variantID.String()})
	rec := httptest.NewRecorder()
	h.DeleteVariant(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteVariant_NotFound(t *testing.T) {
	repo := &stubProductRepo{variant: nil}
	h := newProductHandler(repo)
	req := httptest.NewRequest(http.MethodDelete, "/products/"+uuid.New().String()+"/variants/"+uuid.New().String(), nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParams(req, map[string]string{"id": uuid.New().String(), "variantId": uuid.New().String()})
	rec := httptest.NewRecorder()
	h.DeleteVariant(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteVariant_ModuleDisabled(t *testing.T) {
	h := newProductHandler(&stubProductRepo{})
	req := httptest.NewRequest(http.MethodDelete, "/products/"+uuid.New().String()+"/variants/"+uuid.New().String(), nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), noInventoryTenant()))
	req = withChiParams(req, map[string]string{"id": uuid.New().String(), "variantId": uuid.New().String()})
	rec := httptest.NewRecorder()
	h.DeleteVariant(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

// ── Image CRUD Tests ──────────────────────────────────────────────────────────

func TestAddImage_Valid(t *testing.T) {
	repo := &stubProductRepo{}
	h := newProductHandler(repo)
	productID := uuid.New()
	body, _ := json.Marshal(map[string]any{
		"url":        "https://example.com/images/hero.jpg",
		"sort_order": 0,
		"is_primary": true,
	})
	req := httptest.NewRequest(http.MethodPost, "/products/"+productID.String()+"/images", bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", productID.String())
	rec := httptest.NewRecorder()
	h.AddImage(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAddImage_MissingURL(t *testing.T) {
	h := newProductHandler(&stubProductRepo{})
	body, _ := json.Marshal(map[string]any{"sort_order": 1})
	req := httptest.NewRequest(http.MethodPost, "/products/"+uuid.New().String()+"/images", bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", uuid.New().String())
	rec := httptest.NewRecorder()
	h.AddImage(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAddImage_ProductNotFound(t *testing.T) {
	repo := &stubProductRepo{getErr: errNotFound}
	h := newProductHandler(repo)
	body, _ := json.Marshal(map[string]any{"url": "https://example.com/img.jpg"})
	req := httptest.NewRequest(http.MethodPost, "/products/"+uuid.New().String()+"/images", bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", uuid.New().String())
	rec := httptest.NewRecorder()
	h.AddImage(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAddImage_ModuleDisabled(t *testing.T) {
	h := newProductHandler(&stubProductRepo{})
	body, _ := json.Marshal(map[string]any{"url": "https://example.com/img.jpg"})
	req := httptest.NewRequest(http.MethodPost, "/products/"+uuid.New().String()+"/images", bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), noInventoryTenant()))
	req = withChiParam(req, "id", uuid.New().String())
	rec := httptest.NewRecorder()
	h.AddImage(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAddImage_DuplicateSortOrder(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23505"}
	repo := &stubProductRepo{addImageErr: pgErr}
	h := newProductHandler(repo)
	productID := uuid.New()
	body, _ := json.Marshal(map[string]any{"url": "https://example.com/img.jpg", "sort_order": 1})
	req := httptest.NewRequest(http.MethodPost, "/products/"+productID.String()+"/images", bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", productID.String())
	rec := httptest.NewRecorder()
	h.AddImage(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListImages_OK(t *testing.T) {
	repo := &stubProductRepo{}
	h := newProductHandler(repo)
	productID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/products/"+productID.String()+"/images", nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", productID.String())
	rec := httptest.NewRecorder()
	h.ListImages(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListImages_ProductNotFound(t *testing.T) {
	repo := &stubProductRepo{getErr: errNotFound}
	h := newProductHandler(repo)
	req := httptest.NewRequest(http.MethodGet, "/products/"+uuid.New().String()+"/images", nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParam(req, "id", uuid.New().String())
	rec := httptest.NewRecorder()
	h.ListImages(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUpdateImage_Valid(t *testing.T) {
	repo := &stubProductRepo{}
	h := newProductHandler(repo)
	productID := uuid.New()
	imageID := uuid.New()
	body, _ := json.Marshal(map[string]any{
		"url":        "https://example.com/images/updated.jpg",
		"sort_order": 2,
		"is_primary": false,
	})
	req := httptest.NewRequest(http.MethodPut, "/products/"+productID.String()+"/images/"+imageID.String(), bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParams(req, map[string]string{"id": productID.String(), "imageId": imageID.String()})
	rec := httptest.NewRecorder()
	h.UpdateImage(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateImage_ProductNotFound(t *testing.T) {
	repo := &stubProductRepo{getErr: errNotFound}
	h := newProductHandler(repo)
	body, _ := json.Marshal(map[string]any{"url": "https://example.com/img.jpg", "sort_order": 0})
	req := httptest.NewRequest(http.MethodPut, "/products/"+uuid.New().String()+"/images/"+uuid.New().String(), bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParams(req, map[string]string{"id": uuid.New().String(), "imageId": uuid.New().String()})
	rec := httptest.NewRecorder()
	h.UpdateImage(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestUpdateImage_DuplicateSortOrder(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23505"}
	repo := &stubProductRepo{updateImageErr: pgErr}
	h := newProductHandler(repo)
	productID := uuid.New()
	imageID := uuid.New()
	body, _ := json.Marshal(map[string]any{"url": "https://example.com/img.jpg", "sort_order": 1})
	req := httptest.NewRequest(http.MethodPut, "/products/"+productID.String()+"/images/"+imageID.String(), bytes.NewReader(body))
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParams(req, map[string]string{"id": productID.String(), "imageId": imageID.String()})
	rec := httptest.NewRecorder()
	h.UpdateImage(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteImage_Valid(t *testing.T) {
	repo := &stubProductRepo{}
	h := newProductHandler(repo)
	productID := uuid.New()
	imageID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/products/"+productID.String()+"/images/"+imageID.String(), nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParams(req, map[string]string{"id": productID.String(), "imageId": imageID.String()})
	rec := httptest.NewRecorder()
	h.DeleteImage(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteImage_ProductNotFound(t *testing.T) {
	repo := &stubProductRepo{getErr: errNotFound}
	h := newProductHandler(repo)
	req := httptest.NewRequest(http.MethodDelete, "/products/"+uuid.New().String()+"/images/"+uuid.New().String(), nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParams(req, map[string]string{"id": uuid.New().String(), "imageId": uuid.New().String()})
	rec := httptest.NewRecorder()
	h.DeleteImage(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteImage_InvalidImageID(t *testing.T) {
	h := newProductHandler(&stubProductRepo{})
	req := httptest.NewRequest(http.MethodDelete, "/products/"+uuid.New().String()+"/images/not-a-uuid", nil)
	req = req.WithContext(middleware.WithTenant(req.Context(), activeTenant()))
	req = withChiParams(req, map[string]string{"id": uuid.New().String(), "imageId": "not-a-uuid"})
	rec := httptest.NewRecorder()
	h.DeleteImage(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

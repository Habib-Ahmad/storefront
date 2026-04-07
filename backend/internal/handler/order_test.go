package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/adapter/terminalaf"
	"storefront/backend/internal/db"
	"storefront/backend/internal/handler"
	"storefront/backend/internal/models"
	"storefront/backend/internal/repository"
	"storefront/backend/internal/service"
)

// ── stub order + product repos ────────────────────────────────────────────────

type stubOrderRepo struct{ order *models.Order }

func (s *stubOrderRepo) Create(_ context.Context, o *models.Order, _ []models.OrderItem) error {
	o.ID = uuid.New()
	s.order = o
	return nil
}
func (s *stubOrderRepo) GetByID(_ context.Context, _, _ uuid.UUID) (*models.Order, error) {
	return s.order, nil
}
func (s *stubOrderRepo) GetByIDInternal(_ context.Context, _ uuid.UUID) (*models.Order, error) {
	return s.order, nil
}
func (s *stubOrderRepo) GetByTrackingSlug(_ context.Context, _ string) (*models.Order, error) {
	return s.order, nil
}
func (s *stubOrderRepo) ListByTenant(_ context.Context, _ uuid.UUID, _, _ int) ([]models.Order, error) {
	return nil, nil
}
func (s *stubOrderRepo) CountByTenant(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}
func (s *stubOrderRepo) UpdatePaymentStatus(_ context.Context, _, _ uuid.UUID, _ models.PaymentStatus) error {
	return nil
}
func (s *stubOrderRepo) UpdateFulfillmentStatus(_ context.Context, _, _ uuid.UUID, _ models.FulfillmentStatus) error {
	return nil
}
func (s *stubOrderRepo) ListItems(_ context.Context, _ uuid.UUID) ([]models.OrderItem, error) {
	return nil, nil
}
func (s *stubOrderRepo) WithTx(_ db.DBTX) repository.OrderRepository { return s }

type stubProductRepoForOrder struct{ variant *models.ProductVariant }

func (s *stubProductRepoForOrder) Create(_ context.Context, p *models.Product) error {
	p.ID = uuid.New()
	return nil
}
func (s *stubProductRepoForOrder) GetByID(_ context.Context, _, id uuid.UUID) (*models.Product, error) {
	return &models.Product{ID: id, IsAvailable: true}, nil
}
func (s *stubProductRepoForOrder) ListByTenant(_ context.Context, _ uuid.UUID, _, _ int) ([]models.Product, error) {
	return nil, nil
}
func (s *stubProductRepoForOrder) ListPublicByTenant(_ context.Context, _ uuid.UUID) ([]models.PublicStorefrontProduct, error) {
	return nil, nil
}
func (s *stubProductRepoForOrder) CountByTenant(_ context.Context, _ uuid.UUID) (int, error) {
	return 0, nil
}
func (s *stubProductRepoForOrder) Update(_ context.Context, _ *models.Product) error  { return nil }
func (s *stubProductRepoForOrder) SoftDelete(_ context.Context, _, _ uuid.UUID) error { return nil }
func (s *stubProductRepoForOrder) CreateVariant(_ context.Context, v *models.ProductVariant) error {
	v.ID = uuid.New()
	return nil
}
func (s *stubProductRepoForOrder) GetVariantByID(_ context.Context, _ uuid.UUID) (*models.ProductVariant, error) {
	return s.variant, nil
}
func (s *stubProductRepoForOrder) ListVariants(_ context.Context, _ uuid.UUID) ([]models.ProductVariant, error) {
	return nil, nil
}
func (s *stubProductRepoForOrder) UpdateVariant(_ context.Context, _ *models.ProductVariant) error {
	return nil
}
func (s *stubProductRepoForOrder) SoftDeleteVariant(_ context.Context, _ uuid.UUID) error {
	return nil
}
func (s *stubProductRepoForOrder) DecrementStock(_ context.Context, _ uuid.UUID, _ int) error {
	return nil
}
func (s *stubProductRepoForOrder) RestoreStock(_ context.Context, _ uuid.UUID, _ int) error {
	return nil
}
func (s *stubProductRepoForOrder) AddImage(_ context.Context, _ *models.ProductImage) error {
	return nil
}
func (s *stubProductRepoForOrder) ListImagesByProduct(_ context.Context, _ uuid.UUID) ([]models.ProductImage, error) {
	return nil, nil
}
func (s *stubProductRepoForOrder) DeleteImage(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubProductRepoForOrder) UpdateImage(_ context.Context, _ *models.ProductImage) error {
	return nil
}

type stubPaymentInitiator struct{}

func (s *stubPaymentInitiator) InitiatePayment(_ context.Context, _ *models.Order, _, _ string) (string, error) {
	return "https://paystack.com/pay/stub", nil
}

type stubDispatcher struct{}

func (s *stubDispatcher) Dispatch(_ context.Context, _, _ uuid.UUID, _ terminalaf.BookRequest) (*models.Shipment, error) {
	return &models.Shipment{ID: uuid.New()}, nil
}

type stubTenantRepoForOrder struct{ tenant *models.Tenant }

func (s *stubTenantRepoForOrder) Create(_ context.Context, t *models.Tenant) error {
	t.ID = uuid.New()
	return nil
}
func (s *stubTenantRepoForOrder) GetByID(_ context.Context, _ uuid.UUID) (*models.Tenant, error) {
	return s.tenant, nil
}
func (s *stubTenantRepoForOrder) GetBySlug(_ context.Context, _ string) (*models.Tenant, error) {
	return s.tenant, nil
}
func (s *stubTenantRepoForOrder) Update(_ context.Context, _ *models.Tenant) error { return nil }
func (s *stubTenantRepoForOrder) SoftDelete(_ context.Context, _ uuid.UUID) error  { return nil }
func (s *stubTenantRepoForOrder) WithTx(_ db.DBTX) repository.TenantRepository     { return s }

func newOrderHandler(variant *models.ProductVariant) *handler.OrderHandler {
	svc := service.NewOrderService(&stubOrderRepo{}, &stubProductRepoForOrder{variant: variant})
	return handler.NewOrderHandler(svc, &stubPaymentInitiator{}, &stubDispatcher{}, slog.Default())
}

func newPublicOrderHandler(tenant *models.Tenant, variant *models.ProductVariant) *handler.OrderHandler {
	svc := service.NewOrderService(&stubOrderRepo{}, &stubProductRepoForOrder{variant: variant})
	svc.SetTenantRepo(&stubTenantRepoForOrder{tenant: tenant})
	return handler.NewOrderHandler(svc, &stubPaymentInitiator{}, &stubDispatcher{}, slog.Default())
}

func withURLParam(r *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestCreateOrder_DeliveryMissingPhone(t *testing.T) {
	variantID := uuid.New()
	h := newOrderHandler(&models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(1000), StockQty: nil})
	body, _ := json.Marshal(map[string]any{
		"customer_name":    "Ade",
		"is_delivery":      true,
		"shipping_address": "123 Lagos",
		"items":            []map[string]any{{"variant_id": variantID, "quantity": 1}},
	})
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New(), ActiveModules: models.ActiveModules{Payments: true}}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateOrder_Valid(t *testing.T) {
	variantID := uuid.New()
	h := newOrderHandler(&models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(2500), StockQty: nil})
	phone := "08012345678"
	body, _ := json.Marshal(map[string]any{
		"customer_name":    "Chidi",
		"is_delivery":      true,
		"customer_phone":   phone,
		"shipping_address": "23 Abuja",
		"items":            []map[string]any{{"variant_id": variantID, "quantity": 2}},
	})
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New(), ActiveModules: models.ActiveModules{Payments: true}}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateOrder_CashSale_NoPaystackURL(t *testing.T) {
	variantID := uuid.New()
	h := newOrderHandler(&models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(1500), StockQty: nil})
	body, _ := json.Marshal(map[string]any{
		"payment_method": "cash",
		"items":          []map[string]any{{"variant_id": variantID, "quantity": 1}},
	})
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New(), ActiveModules: models.ActiveModules{Payments: true}}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if url, ok := resp["authorization_url"]; ok && url != "" {
		t.Fatal("cash sale should not return authorization_url")
	}
	if resp["payment_status"] != "paid" {
		t.Fatalf("cash sale should be paid, got %v", resp["payment_status"])
	}
}

func TestCreateOrder_QuickSale_AmountOnly(t *testing.T) {
	h := newOrderHandler(nil)
	body, _ := json.Marshal(map[string]any{
		"payment_method": "cash",
		"total_amount":   3500.00,
		"note":           "walk-in customer",
	})
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New(), ActiveModules: models.ActiveModules{Payments: true}}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["payment_status"] != "paid" {
		t.Fatalf("quick cash sale should be paid, got %v", resp["payment_status"])
	}
}

func TestCreateOrder_QuickSale_MissingAmount(t *testing.T) {
	h := newOrderHandler(nil)
	body, _ := json.Marshal(map[string]any{
		"payment_method": "cash",
	})
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New(), ActiveModules: models.ActiveModules{Payments: true}}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateOrder_InvalidVariantID(t *testing.T) {
	h := newOrderHandler(nil)
	body, _ := json.Marshal(map[string]any{
		"items": []map[string]any{{"variant_id": "not-a-uuid", "quantity": 1}},
	})
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New(), ActiveModules: models.ActiveModules{Payments: true}}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateOrder_ZeroQuantity(t *testing.T) {
	variantID := uuid.New()
	h := newOrderHandler(&models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(1000)})
	body, _ := json.Marshal(map[string]any{
		"items": []map[string]any{{"variant_id": variantID, "quantity": 0}},
	})
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New(), ActiveModules: models.ActiveModules{Payments: true}}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateOrder_InvalidPaymentMethod(t *testing.T) {
	variantID := uuid.New()
	h := newOrderHandler(&models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(1000)})
	body, _ := json.Marshal(map[string]any{
		"payment_method": "bitcoin",
		"items":          []map[string]any{{"variant_id": variantID, "quantity": 1}},
	})
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New(), ActiveModules: models.ActiveModules{Payments: true}}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateOrder_BadJSON(t *testing.T) {
	h := newOrderHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader([]byte("{not json")))
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New(), ActiveModules: models.ActiveModules{Payments: true}}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateOrder_EmptyBody(t *testing.T) {
	h := newOrderHandler(nil)
	body, _ := json.Marshal(map[string]any{})
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New(), ActiveModules: models.ActiveModules{Payments: true}}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreatePublicOrder_Valid(t *testing.T) {
	variantID := uuid.New()
	h := newPublicOrderHandler(&models.Tenant{
		ID:                  uuid.New(),
		Name:                "Funke Fabrics",
		Slug:                "funke-fabrics",
		StorefrontPublished: true,
		Status:              models.TenantStatusActive,
		ActiveModules:       models.ActiveModules{Payments: true},
	}, &models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(2500), StockQty: nil})
	body, _ := json.Marshal(map[string]any{
		"customer_phone":   "08012345678",
		"customer_email":   "chidi@example.com",
		"is_delivery":      true,
		"shipping_address": "23 Abuja",
		"items":            []map[string]any{{"variant_id": variantID, "quantity": 2}},
	})
	req := httptest.NewRequest(http.MethodPost, "/storefronts/funke-fabrics/orders", bytes.NewReader(body))
	req = withURLParam(req, "slug", "funke-fabrics")
	rec := httptest.NewRecorder()
	h.CreatePublic(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Storefront struct {
			Slug string `json:"slug"`
		} `json:"storefront"`
		Order struct {
			TrackingSlug  string `json:"tracking_slug"`
			PaymentStatus string `json:"payment_status"`
		} `json:"order"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Storefront.Slug != "funke-fabrics" {
		t.Fatalf("storefront slug: want funke-fabrics, got %s", resp.Storefront.Slug)
	}
	if resp.Order.PaymentStatus != "pending" {
		t.Fatalf("payment status: want pending, got %s", resp.Order.PaymentStatus)
	}
	if resp.Order.TrackingSlug == "" {
		t.Fatal("expected tracking slug in response")
	}
}

func TestCreatePublicOrder_CheckoutUnavailable(t *testing.T) {
	variantID := uuid.New()
	h := newPublicOrderHandler(&models.Tenant{
		ID:                  uuid.New(),
		Name:                "Funke Fabrics",
		Slug:                "funke-fabrics",
		StorefrontPublished: true,
		Status:              models.TenantStatusActive,
		ActiveModules:       models.ActiveModules{Payments: false},
	}, &models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(2500), StockQty: nil})
	body, _ := json.Marshal(map[string]any{
		"customer_name":  "Chidi",
		"customer_phone": "08012345678",
		"items":          []map[string]any{{"variant_id": variantID, "quantity": 1}},
	})
	req := httptest.NewRequest(http.MethodPost, "/storefronts/funke-fabrics/orders", bytes.NewReader(body))
	req = withURLParam(req, "slug", "funke-fabrics")
	rec := httptest.NewRecorder()
	h.CreatePublic(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

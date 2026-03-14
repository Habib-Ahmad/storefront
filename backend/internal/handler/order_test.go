package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/adapter/terminalaf"
	"storefront/backend/internal/handler"
	"storefront/backend/internal/models"
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

func newOrderHandler(variant *models.ProductVariant) *handler.OrderHandler {
	svc := service.NewOrderService(&stubOrderRepo{}, &stubProductRepoForOrder{variant: variant})
	return handler.NewOrderHandler(svc, &stubPaymentInitiator{}, &stubDispatcher{}, slog.Default())
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
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New()}))
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
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New()}))
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
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New()}))
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
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New()}))
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
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New()}))
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
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New()}))
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
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New()}))
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
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New()}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateOrder_BadJSON(t *testing.T) {
	h := newOrderHandler(nil)
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader([]byte("{not json")))
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New()}))
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
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New()}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

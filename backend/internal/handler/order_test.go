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
func (s *stubOrderRepo) GetByID(_ context.Context, _ uuid.UUID) (*models.Order, error) {
	return s.order, nil
}
func (s *stubOrderRepo) GetByTrackingSlug(_ context.Context, _ string) (*models.Order, error) {
	return s.order, nil
}
func (s *stubOrderRepo) ListByTenant(_ context.Context, _ uuid.UUID, _, _ int) ([]models.Order, error) {
	return nil, nil
}
func (s *stubOrderRepo) UpdatePaymentStatus(_ context.Context, _ uuid.UUID, _ models.PaymentStatus) error {
	return nil
}
func (s *stubOrderRepo) UpdateFulfillmentStatus(_ context.Context, _ uuid.UUID, _ models.FulfillmentStatus) error {
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
func (s *stubProductRepoForOrder) GetByID(_ context.Context, _ uuid.UUID) (*models.Product, error) {
	return nil, nil
}
func (s *stubProductRepoForOrder) ListByTenant(_ context.Context, _ uuid.UUID) ([]models.Product, error) {
	return nil, nil
}
func (s *stubProductRepoForOrder) Update(_ context.Context, _ *models.Product) error { return nil }
func (s *stubProductRepoForOrder) SoftDelete(_ context.Context, _ uuid.UUID) error   { return nil }
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

type stubPaymentInitiator struct{}

func (s *stubPaymentInitiator) InitiatePayment(_ context.Context, _ *models.Order, _, _ string) (string, error) {
	return "https://paystack.com/pay/stub", nil
}

func newOrderHandler(variant *models.ProductVariant) *handler.OrderHandler {
	svc := service.NewOrderService(&stubOrderRepo{}, &stubProductRepoForOrder{variant: variant})
	return handler.NewOrderHandler(svc, &stubPaymentInitiator{}, slog.Default())
}

func TestCreateOrder_MissingCustomerName(t *testing.T) {
	h := newOrderHandler(nil)
	body, _ := json.Marshal(map[string]any{"is_delivery": false})
	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewReader(body))
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New()}))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
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

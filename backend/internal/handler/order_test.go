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
	"github.com/jackc/pgx/v5"
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
func (s *stubOrderRepo) GetByPublicCheckoutID(_ context.Context, _ uuid.UUID, checkoutID uuid.UUID) (*models.Order, error) {
	if s.order != nil && s.order.PublicCheckoutID != nil && *s.order.PublicCheckoutID == checkoutID && s.order.PaymentStatus != models.PaymentStatusFailed {
		return s.order, nil
	}
	return nil, pgx.ErrNoRows
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
func (s *stubProductRepoForOrder) WithTx(_ db.DBTX) repository.ProductRepository { return s }

type stubPaymentInitiator struct {
	authorizationURL string
	customerEmail    string
	subaccountCode   string
	callbackURL      string
	initErr          error
	successErr       error
	successReference string
	failedReference  string
}

func (s *stubPaymentInitiator) InitiatePayment(_ context.Context, _ *models.Order, customerEmail, subaccountCode, callbackURL string) (string, error) {
	s.customerEmail = customerEmail
	s.subaccountCode = subaccountCode
	s.callbackURL = callbackURL
	if s.initErr != nil {
		return "", s.initErr
	}
	if s.authorizationURL != "" {
		return s.authorizationURL, nil
	}
	return "https://paystack.com/pay/stub", nil
}

func (s *stubPaymentInitiator) HandleChargeSuccess(_ context.Context, reference string) error {
	s.successReference = reference
	return s.successErr
}

func (s *stubPaymentInitiator) HandleChargeFailed(_ context.Context, reference string) error {
	s.failedReference = reference
	return nil
}

type stubDispatcher struct{}

func (s *stubDispatcher) Dispatch(_ context.Context, _, _ uuid.UUID, _ terminalaf.BookRequest) (*models.Shipment, error) {
	return &models.Shipment{ID: uuid.New()}, nil
}

type stubDeliveryQuoter struct {
	shippingFee          decimal.Decimal
	err                  error
	quoteResponse        *models.PublicStorefrontDeliveryQuoteResponse
	quoteCalls           int
	resolveCalls         int
	lastQuoteSlug        string
	lastQuoteRequest     models.PublicStorefrontDeliveryQuoteRequest
	lastResolveSlug      string
	lastResolveRequest   models.PublicStorefrontDeliveryQuoteRequest
	lastResolveSelection models.PublicStorefrontDeliveryQuoteSelection
}

func (s *stubDeliveryQuoter) QuotePublic(_ context.Context, _ string, _ models.PublicStorefrontDeliveryQuoteRequest) (*models.PublicStorefrontDeliveryQuoteResponse, error) {
	s.quoteCalls++
	if s.err != nil {
		return nil, s.err
	}
	if s.quoteResponse != nil {
		return s.quoteResponse, nil
	}
	return &models.PublicStorefrontDeliveryQuoteResponse{}, nil
}

func (s *stubDeliveryQuoter) ResolvePublicSelection(_ context.Context, slug string, req models.PublicStorefrontDeliveryQuoteRequest, selection models.PublicStorefrontDeliveryQuoteSelection) (decimal.Decimal, error) {
	s.resolveCalls++
	s.lastResolveSlug = slug
	s.lastResolveRequest = req
	s.lastResolveSelection = selection
	if s.err != nil {
		return decimal.Zero, s.err
	}
	if s.shippingFee.IsZero() {
		return decimal.NewFromInt(1500), nil
	}
	return s.shippingFee, nil
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
	return newOrderHandlerWithPayment(variant, &stubPaymentInitiator{})
}

func newOrderHandlerWithPayment(variant *models.ProductVariant, payment *stubPaymentInitiator) *handler.OrderHandler {
	svc := service.NewOrderService(&stubOrderRepo{}, &stubProductRepoForOrder{variant: variant})
	return handler.NewOrderHandler(svc, payment, &stubDispatcher{}, "", slog.Default())
}

func newPublicOrderHandler(tenant *models.Tenant, variant *models.ProductVariant) *handler.OrderHandler {
	return newPublicOrderHandlerWithPayment(tenant, variant, &stubPaymentInitiator{})
}

func newPublicOrderHandlerWithPayment(tenant *models.Tenant, variant *models.ProductVariant, payment *stubPaymentInitiator) *handler.OrderHandler {
	svc := service.NewOrderService(&stubOrderRepo{}, &stubProductRepoForOrder{variant: variant})
	svc.SetTenantRepo(&stubTenantRepoForOrder{tenant: tenant})
	h := handler.NewOrderHandler(svc, payment, &stubDispatcher{}, "https://storefront.test", slog.Default())
	h.SetDeliveryQuoteService(&stubDeliveryQuoter{shippingFee: decimal.NewFromInt(1500)})
	return h
}

func newPublicOrderHandlerWithServices(tenant *models.Tenant, variant *models.ProductVariant, payment *stubPaymentInitiator, quoter *stubDeliveryQuoter) *handler.OrderHandler {
	svc := service.NewOrderService(&stubOrderRepo{}, &stubProductRepoForOrder{variant: variant})
	svc.SetTenantRepo(&stubTenantRepoForOrder{tenant: tenant})
	h := handler.NewOrderHandler(svc, payment, &stubDispatcher{}, "https://storefront.test", slog.Default())
	h.SetDeliveryQuoteService(quoter)
	return h
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
	payment := &stubPaymentInitiator{}
	quoter := &stubDeliveryQuoter{shippingFee: decimal.NewFromInt(1500)}
	h := newPublicOrderHandlerWithServices(&models.Tenant{
		ID:                  uuid.New(),
		Name:                "Funke Fabrics",
		Slug:                "funke-fabrics",
		StorefrontPublished: true,
		Status:              models.TenantStatusActive,
		ActiveModules:       models.ActiveModules{Payments: true, Logistics: true},
	}, &models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(2500), StockQty: nil}, payment, quoter)
	body, _ := json.Marshal(map[string]any{
		"checkout_id":      uuid.New().String(),
		"customer_phone":   "08012345678",
		"is_delivery":      true,
		"shipping_address": "23 Abuja",
		"delivery_option":  map[string]any{"courier_id": "123", "service_code": "bike"},
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
			ShippingFee   string `json:"shipping_fee"`
		} `json:"order"`
		AuthorizationURL string `json:"authorization_url"`
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
	if resp.Order.ShippingFee != "1500" {
		t.Fatalf("shipping fee: want 1500, got %s", resp.Order.ShippingFee)
	}
	if quoter.lastResolveRequest.CustomerName != "Guest customer" {
		t.Fatalf("expected fallback customer name, got %q", quoter.lastResolveRequest.CustomerName)
	}
	if resp.Order.TrackingSlug == "" {
		t.Fatal("expected tracking slug in response")
	}
	if resp.AuthorizationURL == "" {
		t.Fatal("expected authorization_url in response")
	}
	if payment.customerEmail != "guest@storefront.ng" {
		t.Fatalf("expected guest email fallback, got %s", payment.customerEmail)
	}
	if payment.callbackURL != "https://storefront.test/order/"+resp.Order.TrackingSlug {
		t.Fatalf("unexpected callback URL: %s", payment.callbackURL)
	}
}

func TestCreatePublicOrder_DeliveryRequiresSelectedQuote(t *testing.T) {
	variantID := uuid.New()
	h := newPublicOrderHandler(&models.Tenant{
		ID:                  uuid.New(),
		Name:                "Funke Fabrics",
		Slug:                "funke-fabrics",
		StorefrontPublished: true,
		Status:              models.TenantStatusActive,
		ActiveModules:       models.ActiveModules{Payments: true, Logistics: true},
	}, &models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(2500), StockQty: nil})
	body, _ := json.Marshal(map[string]any{
		"checkout_id":      uuid.New().String(),
		"customer_phone":   "08012345678",
		"is_delivery":      true,
		"shipping_address": "23 Abuja",
		"items":            []map[string]any{{"variant_id": variantID, "quantity": 2}},
	})
	req := httptest.NewRequest(http.MethodPost, "/storefronts/funke-fabrics/orders", bytes.NewReader(body))
	req = withURLParam(req, "slug", "funke-fabrics")
	rec := httptest.NewRecorder()

	h.CreatePublic(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestQuotePublicDelivery_Valid(t *testing.T) {
	variantID := uuid.New()
	quoter := &stubDeliveryQuoter{
		quoteResponse: &models.PublicStorefrontDeliveryQuoteResponse{
			Storefront: models.PublicStorefront{Slug: "funke-fabrics", Name: "Funke Fabrics"},
			Options: []models.PublicStorefrontDeliveryQuoteOption{{
				ID:          "123:bike:dropoff",
				CourierID:   "123",
				CourierName: "Kwik",
				ServiceCode: "bike",
				ServiceType: "dropoff",
				Amount:      decimal.NewFromInt(3500),
				Currency:    "NGN",
			}},
		},
	}
	h := newPublicOrderHandlerWithServices(&models.Tenant{
		ID:                  uuid.New(),
		Name:                "Funke Fabrics",
		Slug:                "funke-fabrics",
		StorefrontPublished: true,
		Status:              models.TenantStatusActive,
		ActiveModules:       models.ActiveModules{Payments: true, Logistics: true},
	}, &models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(2500), StockQty: nil}, &stubPaymentInitiator{}, quoter)
	body, _ := json.Marshal(models.PublicStorefrontDeliveryQuoteRequest{
		CustomerName:    "Chidi",
		CustomerPhone:   "08012345678",
		ShippingAddress: "23 Abuja",
		Items:           []models.PublicStorefrontDeliveryQuoteRequestItem{{VariantID: variantID, Quantity: 2}},
	})
	req := httptest.NewRequest(http.MethodPost, "/storefronts/funke-fabrics/delivery-quotes", bytes.NewReader(body))
	req = withURLParam(req, "slug", "funke-fabrics")
	rec := httptest.NewRecorder()

	h.QuotePublicDelivery(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp models.PublicStorefrontDeliveryQuoteResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Options) != 1 || resp.Options[0].CourierName != "Kwik" {
		t.Fatalf("unexpected quote response: %+v", resp.Options)
	}
}

func TestCreateOrder_InitPaymentFailureReturnsError(t *testing.T) {
	variantID := uuid.New()
	payment := &stubPaymentInitiator{initErr: errors.New("paystack unavailable")}
	h := newOrderHandlerWithPayment(&models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(2500), StockQty: nil}, payment)
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
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", rec.Code, rec.Body.String())
	}
	if payment.failedReference == "" {
		t.Fatal("expected payment rollback to be triggered")
	}
}

func TestResumePayment_Valid(t *testing.T) {
	tenantID := uuid.New()
	orderID := uuid.New()
	subaccount := "SUB_merchant"
	payment := &stubPaymentInitiator{authorizationURL: "https://paystack.com/pay/resume"}
	svc := service.NewOrderService(&stubOrderRepo{order: &models.Order{
		ID:                orderID,
		TenantID:          tenantID,
		PaymentMethod:     models.PaymentMethodOnline,
		PaymentStatus:     models.PaymentStatusPending,
		FulfillmentStatus: models.FulfillmentStatusProcessing,
	}}, &stubProductRepoForOrder{})
	h := handler.NewOrderHandler(svc, payment, &stubDispatcher{}, "", slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/orders/"+orderID.String()+"/resume-payment", nil)
	req = withURLParam(req, "id", orderID.String())
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: tenantID, PaystackSubaccountID: &subaccount}))
	rec := httptest.NewRecorder()

	h.ResumePayment(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		AuthorizationURL string `json:"authorization_url"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.AuthorizationURL != "https://paystack.com/pay/resume" {
		t.Fatalf("unexpected authorization URL: %s", resp.AuthorizationURL)
	}
	if payment.customerEmail != "guest@storefront.ng" {
		t.Fatalf("expected guest email fallback, got %s", payment.customerEmail)
	}
	if payment.subaccountCode != subaccount {
		t.Fatalf("expected subaccount %s, got %s", subaccount, payment.subaccountCode)
	}
	if payment.callbackURL != "" {
		t.Fatalf("merchant resume should not set callback URL, got %s", payment.callbackURL)
	}
}

func TestResumePaymentPublic_Valid(t *testing.T) {
	tenantID := uuid.New()
	payment := &stubPaymentInitiator{authorizationURL: "https://paystack.com/pay/public-resume"}
	subaccount := "SUB_public"
	trackingSlug := "abc123def456"
	svc := service.NewOrderService(&stubOrderRepo{order: &models.Order{
		ID:                uuid.New(),
		TenantID:          tenantID,
		TrackingSlug:      trackingSlug,
		PaymentMethod:     models.PaymentMethodOnline,
		PaymentStatus:     models.PaymentStatusPending,
		FulfillmentStatus: models.FulfillmentStatusProcessing,
	}}, &stubProductRepoForOrder{})
	svc.SetTenantRepo(&stubTenantRepoForOrder{tenant: &models.Tenant{ID: tenantID, PaystackSubaccountID: &subaccount}})
	h := handler.NewOrderHandler(svc, payment, &stubDispatcher{}, "https://storefront.test", slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/track/"+trackingSlug+"/resume-payment", nil)
	req = withURLParam(req, "slug", trackingSlug)
	rec := httptest.NewRecorder()

	h.ResumePaymentPublic(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if payment.callbackURL != "https://storefront.test/order/"+trackingSlug {
		t.Fatalf("unexpected callback URL: %s", payment.callbackURL)
	}
	if payment.subaccountCode != subaccount {
		t.Fatalf("expected subaccount %s, got %s", subaccount, payment.subaccountCode)
	}
}

func TestConfirmPaymentPublic_Valid(t *testing.T) {
	tenantID := uuid.New()
	orderID := uuid.New()
	trackingSlug := "abc123def456"
	payment := &stubPaymentInitiator{}
	svc := service.NewOrderService(&stubOrderRepo{order: &models.Order{
		ID:                orderID,
		TenantID:          tenantID,
		TrackingSlug:      trackingSlug,
		PaymentMethod:     models.PaymentMethodOnline,
		PaymentStatus:     models.PaymentStatusPending,
		FulfillmentStatus: models.FulfillmentStatusProcessing,
	}}, &stubProductRepoForOrder{})
	h := handler.NewOrderHandler(svc, payment, &stubDispatcher{}, "https://storefront.test", slog.Default())
	body, _ := json.Marshal(map[string]any{"reference": orderID.String()})
	req := httptest.NewRequest(http.MethodPost, "/track/"+trackingSlug+"/confirm-payment", bytes.NewReader(body))
	req = withURLParam(req, "slug", trackingSlug)
	rec := httptest.NewRecorder()

	h.ConfirmPaymentPublic(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if payment.successReference != orderID.String() {
		t.Fatalf("expected payment confirmation for %s, got %s", orderID, payment.successReference)
	}
}

func TestConfirmPaymentPublic_RejectsMismatchedReference(t *testing.T) {
	tenantID := uuid.New()
	trackingSlug := "abc123def456"
	payment := &stubPaymentInitiator{}
	svc := service.NewOrderService(&stubOrderRepo{order: &models.Order{
		ID:                uuid.New(),
		TenantID:          tenantID,
		TrackingSlug:      trackingSlug,
		PaymentMethod:     models.PaymentMethodOnline,
		PaymentStatus:     models.PaymentStatusPending,
		FulfillmentStatus: models.FulfillmentStatusProcessing,
	}}, &stubProductRepoForOrder{})
	h := handler.NewOrderHandler(svc, payment, &stubDispatcher{}, "https://storefront.test", slog.Default())
	body, _ := json.Marshal(map[string]any{"reference": uuid.New().String()})
	req := httptest.NewRequest(http.MethodPost, "/track/"+trackingSlug+"/confirm-payment", bytes.NewReader(body))
	req = withURLParam(req, "slug", trackingSlug)
	rec := httptest.NewRecorder()

	h.ConfirmPaymentPublic(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestResumePayment_RejectsCancelledOrder(t *testing.T) {
	tenantID := uuid.New()
	orderID := uuid.New()
	payment := &stubPaymentInitiator{}
	svc := service.NewOrderService(&stubOrderRepo{order: &models.Order{
		ID:                orderID,
		TenantID:          tenantID,
		PaymentMethod:     models.PaymentMethodOnline,
		PaymentStatus:     models.PaymentStatusFailed,
		FulfillmentStatus: models.FulfillmentStatusCancelled,
	}}, &stubProductRepoForOrder{})
	h := handler.NewOrderHandler(svc, payment, &stubDispatcher{}, "", slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/orders/"+orderID.String()+"/resume-payment", nil)
	req = withURLParam(req, "id", orderID.String())
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: tenantID}))
	rec := httptest.NewRecorder()

	h.ResumePayment(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreatePublicOrder_InitPaymentFailureReturnsError(t *testing.T) {
	variantID := uuid.New()
	payment := &stubPaymentInitiator{initErr: errors.New("paystack unavailable")}
	h := newPublicOrderHandlerWithPayment(&models.Tenant{
		ID:                  uuid.New(),
		Name:                "Funke Fabrics",
		Slug:                "funke-fabrics",
		StorefrontPublished: true,
		Status:              models.TenantStatusActive,
		ActiveModules:       models.ActiveModules{Payments: true, Logistics: true},
	}, &models.ProductVariant{ID: variantID, Price: decimal.NewFromInt(2500), StockQty: nil}, payment)
	body, _ := json.Marshal(map[string]any{
		"checkout_id":      uuid.New().String(),
		"customer_phone":   "08012345678",
		"is_delivery":      true,
		"shipping_address": "23 Abuja",
		"delivery_option":  map[string]any{"courier_id": "123", "service_code": "bike"},
		"items":            []map[string]any{{"variant_id": variantID, "quantity": 2}},
	})
	req := httptest.NewRequest(http.MethodPost, "/storefronts/funke-fabrics/orders", bytes.NewReader(body))
	req = withURLParam(req, "slug", "funke-fabrics")
	rec := httptest.NewRecorder()
	h.CreatePublic(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", rec.Code, rec.Body.String())
	}
	if payment.failedReference == "" {
		t.Fatal("expected payment rollback to be triggered")
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
		"checkout_id":    uuid.New().String(),
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

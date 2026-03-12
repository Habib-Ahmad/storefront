import os

BASE = "/Users/ahmad.habib/Desktop/projects/storefront/backend"

def write(rel_path, content):
    full = os.path.join(BASE, rel_path)
    os.makedirs(os.path.dirname(full), exist_ok=True)
    with open(full, 'w') as f:
        f.write(content)
    pkgs = [l for l in content.split('\n') if l.startswith('package ')]
    status = "OK" if len(pkgs) == 1 else f"BAD({len(pkgs)} package decls)"
    print(f"{status}: {rel_path}")

write("internal/service/payment.go", """package service

import (
\t"context"
\t"errors"
\t"fmt"

\t"github.com/google/uuid"

\tpaystack "storefront/backend/internal/adapter/paystack"
\t"storefront/backend/internal/models"
\t"storefront/backend/internal/repository"
)

var ErrPaymentVerificationFailed = errors.New("payment verification failed")

// PaystackClient is the subset of paystack.Client used by PaymentService.
type PaystackClient interface {
\tInitializeTransaction(ctx context.Context, req paystack.InitializeRequest) (*paystack.InitializeResponse, error)
\tVerifyTransaction(ctx context.Context, reference string) (*paystack.VerifyResponse, error)
}

type PaymentService struct {
\tpaystack  PaystackClient
\torders    repository.OrderRepository
\twalletSvc *WalletService
}

func NewPaymentService(
\tpaystackClient PaystackClient,
\torders repository.OrderRepository,
\twalletSvc *WalletService,
) *PaymentService {
\treturn &PaymentService{paystack: paystackClient, orders: orders, walletSvc: walletSvc}
}

// InitiatePayment creates a Paystack transaction and returns the redirect URL.
// The Paystack reference equals the order UUID so webhook callbacks can reverse-map it.
func (s *PaymentService) InitiatePayment(ctx context.Context, order *models.Order, customerEmail, subaccountCode string) (string, error) {
\treq := paystack.InitializeRequest{
\t\tEmail:          customerEmail,
\t\tAmount:         order.TotalAmount.Add(order.ShippingFee),
\t\tReference:      order.ID.String(),
\t\tSubaccountCode: subaccountCode,
\t\tMetadata:       map[string]any{"order_id": order.ID},
\t}
\tresp, err := s.paystack.InitializeTransaction(ctx, req)
\tif err != nil {
\t\treturn "", fmt.Errorf("initiate payment: %w", err)
\t}
\treturn resp.AuthorizationURL, nil
}

// HandleChargeSuccess verifies the Paystack charge event and credits the tenant wallet.
// reference is the order UUID string used as Paystack reference at initialization.
func (s *PaymentService) HandleChargeSuccess(ctx context.Context, reference string) error {
\tresp, err := s.paystack.VerifyTransaction(ctx, reference)
\tif err != nil {
\t\treturn fmt.Errorf("verify transaction: %w", err)
\t}
\tif resp.Status != "success" {
\t\treturn ErrPaymentVerificationFailed
\t}

\torderID, err := uuid.Parse(reference)
\tif err != nil {
\t\treturn fmt.Errorf("invalid reference: %w", err)
\t}

\torder, err := s.orders.GetByID(ctx, orderID)
\tif err != nil {
\t\treturn fmt.Errorf("get order: %w", err)
\t}

\tif err := s.orders.UpdatePaymentStatus(ctx, orderID, models.PaymentStatusPaid); err != nil {
\t\treturn fmt.Errorf("update payment status: %w", err)
\t}

\t// walletSvc.Credit looks up the wallet by tenant ID.
\tif _, err := s.walletSvc.Credit(ctx, order.TenantID, resp.Amount, &orderID); err != nil {
\t\treturn fmt.Errorf("credit wallet: %w", err)
\t}

\treturn nil
}
""")

write("internal/service/shipment.go", """package service

import (
\t"context"
\t"encoding/json"
\t"fmt"

\t"github.com/google/uuid"

\t"storefront/backend/internal/adapter/terminalaf"
\t"storefront/backend/internal/models"
\t"storefront/backend/internal/repository"
)

// CarrierClient abstracts the logistics provider (Terminal Africa primary; swap point for Shipbubble).
type CarrierClient interface {
\tBookShipment(ctx context.Context, req terminalaf.BookRequest) (*terminalaf.BookResponse, error)
}

type ShipmentService struct {
\tcarrier   CarrierClient
\tshipments repository.ShipmentRepository
\torders    repository.OrderRepository
\twalletSvc *WalletService
}

func NewShipmentService(
\tcarrier CarrierClient,
\tshipments repository.ShipmentRepository,
\torders repository.OrderRepository,
\twalletSvc *WalletService,
) *ShipmentService {
\treturn &ShipmentService{carrier: carrier, shipments: shipments, orders: orders, walletSvc: walletSvc}
}

// Dispatch books a shipment with the carrier and persists the booking to the shipments table.
func (s *ShipmentService) Dispatch(ctx context.Context, orderID, tenantID uuid.UUID, req terminalaf.BookRequest) (*models.Shipment, error) {
\tresp, err := s.carrier.BookShipment(ctx, req)
\tif err != nil {
\t\treturn nil, fmt.Errorf("book shipment: %w", err)
\t}

\thistory, _ := json.Marshal(resp)
\tshipment := &models.Shipment{
\t\tOrderID:        orderID,
\t\tTenantID:       tenantID,
\t\tCarrierRef:     &resp.CarrierRef,
\t\tTrackingNumber: &resp.TrackingNumber,
\t\tCarrierHistory: history,
\t}
\tif err := s.shipments.Create(ctx, shipment); err != nil {
\t\treturn nil, fmt.Errorf("save shipment: %w", err)
\t}

\tif err := s.orders.UpdateFulfillmentStatus(ctx, orderID, models.FulfillmentStatusShipped); err != nil {
\t\treturn nil, fmt.Errorf("update fulfillment status: %w", err)
\t}

\treturn shipment, nil
}

// HandleDelivered processes a Terminal Africa delivery webhook:
// updates shipment + order status, then releases the order amount from pending balance.
// orderID is extracted from the booking metadata_reference field echoed in the webhook.
func (s *ShipmentService) HandleDelivered(ctx context.Context, orderID uuid.UUID) error {
\torder, err := s.orders.GetByID(ctx, orderID)
\tif err != nil {
\t\treturn fmt.Errorf("get order: %w", err)
\t}

\tshipment, err := s.shipments.GetByOrderID(ctx, orderID)
\tif err != nil {
\t\treturn fmt.Errorf("get shipment: %w", err)
\t}

\tif err := s.shipments.UpdateStatus(ctx, shipment.ID, models.ShipmentStatusDelivered); err != nil {
\t\treturn fmt.Errorf("update shipment status: %w", err)
\t}

\tif err := s.orders.UpdateFulfillmentStatus(ctx, orderID, models.FulfillmentStatusDelivered); err != nil {
\t\treturn fmt.Errorf("update fulfillment status: %w", err)
\t}

\t// Release total order value (total + shipping) from pending to available balance.
\tamount := order.TotalAmount.Add(order.ShippingFee)
\tif err := s.walletSvc.ReleasePending(ctx, order.TenantID, amount); err != nil {
\t\treturn fmt.Errorf("release pending: %w", err)
\t}

\treturn nil
}
""")

write("internal/service/payment_test.go", """package service_test

import (
\t"context"
\t"errors"
\t"testing"

\t"github.com/google/uuid"
\t"github.com/shopspring/decimal"

\tpaystack "storefront/backend/internal/adapter/paystack"
\t"storefront/backend/internal/models"
\t"storefront/backend/internal/service"
)

// ── Paystack client mock ──────────────────────────────────────────────────────

type mockPaystackClient struct {
\tinitResp *paystack.InitializeResponse
\tinitErr  error
\tverResp  *paystack.VerifyResponse
\tverErr   error
}

func (m *mockPaystackClient) InitializeTransaction(_ context.Context, _ paystack.InitializeRequest) (*paystack.InitializeResponse, error) {
\treturn m.initResp, m.initErr
}

func (m *mockPaystackClient) VerifyTransaction(_ context.Context, _ string) (*paystack.VerifyResponse, error) {
\treturn m.verResp, m.verErr
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newPaymentSvc(ps *mockPaystackClient, orders *mockOrderRepo, wallet *models.Wallet) *service.PaymentService {
\twalletSvc := service.NewWalletService(
\t\t&mockWalletRepo{wallet: wallet},
\t\t&mockTxRepo{},
\t\t&mockTenantRepo{},
\t\ttestHMACSecret,
\t)
\treturn service.NewPaymentService(ps, orders, walletSvc)
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestInitiatePayment_ReturnsAuthURL(t *testing.T) {
\torderID := uuid.New()
\torder := &models.Order{
\t\tID:          orderID,
\t\tTotalAmount: decimal.NewFromInt(5000),
\t\tShippingFee: decimal.NewFromInt(500),
\t}
\tps := &mockPaystackClient{
\t\tinitResp: &paystack.InitializeResponse{AuthorizationURL: "https://paystack.com/pay/xyz"},
\t}
\tsvc := newPaymentSvc(ps, &mockOrderRepo{order: order}, &models.Wallet{ID: uuid.New()})

\turl, err := svc.InitiatePayment(context.Background(), order, "buyer@example.com", "SUB_123")
\tif err != nil {
\t\tt.Fatalf("unexpected error: %v", err)
\t}
\tif url != "https://paystack.com/pay/xyz" {
\t\tt.Fatalf("expected auth URL, got %s", url)
\t}
}

func TestInitiatePayment_AdapterError(t *testing.T) {
\torder := &models.Order{ID: uuid.New()}
\tps := &mockPaystackClient{initErr: errors.New("paystack down")}
\tsvc := newPaymentSvc(ps, &mockOrderRepo{order: order}, &models.Wallet{ID: uuid.New()})

\t_, err := svc.InitiatePayment(context.Background(), order, "buyer@example.com", "")
\tif err == nil {
\t\tt.Fatal("expected error, got nil")
\t}
}

func TestHandleChargeSuccess_CreditsWallet(t *testing.T) {
\ttenantID := uuid.New()
\torderID := uuid.New()
\torder := &models.Order{ID: orderID, TenantID: tenantID, TotalAmount: decimal.NewFromInt(5000)}
\twallet := &models.Wallet{ID: uuid.New(), TenantID: tenantID}

\tps := &mockPaystackClient{
\t\tverResp: &paystack.VerifyResponse{
\t\t\tStatus:    "success",
\t\t\tAmount:    decimal.NewFromInt(5000),
\t\t\tReference: orderID.String(),
\t\t},
\t}
\ttxRepo := &mockTxRepo{}
\twalletSvc := service.NewWalletService(
\t\t&mockWalletRepo{wallet: wallet},
\t\ttxRepo,
\t\t&mockTenantRepo{},
\t\ttestHMACSecret,
\t)
\tpaymentSvc := service.NewPaymentService(ps, &mockOrderRepo{order: order}, walletSvc)

\terr := paymentSvc.HandleChargeSuccess(context.Background(), orderID.String())
\tif err != nil {
\t\tt.Fatalf("unexpected error: %v", err)
\t}
\tif txRepo.created == nil {
\t\tt.Fatal("expected wallet credit transaction to be created")
\t}
}

func TestHandleChargeSuccess_VerificationFails(t *testing.T) {
\tps := &mockPaystackClient{
\t\tverResp: &paystack.VerifyResponse{Status: "failed"},
\t}
\tsvc := newPaymentSvc(ps, &mockOrderRepo{order: &models.Order{ID: uuid.New()}}, &models.Wallet{ID: uuid.New()})

\terr := svc.HandleChargeSuccess(context.Background(), uuid.New().String())
\tif !errors.Is(err, service.ErrPaymentVerificationFailed) {
\t\tt.Fatalf("expected ErrPaymentVerificationFailed, got %v", err)
\t}
}

func TestHandleChargeSuccess_InvalidReference(t *testing.T) {
\tps := &mockPaystackClient{
\t\tverResp: &paystack.VerifyResponse{Status: "success"},
\t}
\tsvc := newPaymentSvc(ps, &mockOrderRepo{}, &models.Wallet{ID: uuid.New()})

\terr := svc.HandleChargeSuccess(context.Background(), "not-a-uuid")
\tif err == nil {
\t\tt.Fatal("expected error for invalid reference, got nil")
\t}
}
""")

write("internal/service/shipment_test.go", """package service_test

import (
\t"context"
\t"errors"
\t"testing"

\t"github.com/google/uuid"
\t"github.com/shopspring/decimal"

\t"storefront/backend/internal/adapter/terminalaf"
\t"storefront/backend/internal/models"
\t"storefront/backend/internal/service"
)

// ── carrier client mock ───────────────────────────────────────────────────────

type mockCarrierClient struct {
\tresp *terminalaf.BookResponse
\terr  error
}

func (m *mockCarrierClient) BookShipment(_ context.Context, _ terminalaf.BookRequest) (*terminalaf.BookResponse, error) {
\treturn m.resp, m.err
}

// ── shipment repo mock ────────────────────────────────────────────────────────

type mockShipmentRepo struct {
\tshipment  *models.Shipment
\terr       error
\tstatusSet models.ShipmentStatus
}

func (m *mockShipmentRepo) Create(_ context.Context, s *models.Shipment) error {
\ts.ID = uuid.New()
\tm.shipment = s
\treturn m.err
}

func (m *mockShipmentRepo) GetByOrderID(_ context.Context, _ uuid.UUID) (*models.Shipment, error) {
\treturn m.shipment, m.err
}

func (m *mockShipmentRepo) UpdateStatus(_ context.Context, _ uuid.UUID, status models.ShipmentStatus) error {
\tm.statusSet = status
\treturn m.err
}

func (m *mockShipmentRepo) AppendCarrierEvent(_ context.Context, _ uuid.UUID, _ []byte) error {
\treturn m.err
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newShipmentSvc(carrier *mockCarrierClient, shipments *mockShipmentRepo, orders *mockOrderRepo, wallet *models.Wallet) *service.ShipmentService {
\twalletSvc := service.NewWalletService(
\t\t&mockWalletRepo{wallet: wallet},
\t\t&mockTxRepo{},
\t\t&mockTenantRepo{},
\t\ttestHMACSecret,
\t)
\treturn service.NewShipmentService(carrier, shipments, orders, walletSvc)
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestDispatch_BookAndSave(t *testing.T) {
\torderID := uuid.New()
\ttenantID := uuid.New()
\tcarrier := &mockCarrierClient{
\t\tresp: &terminalaf.BookResponse{CarrierRef: "TA-001", TrackingNumber: "TRK-001", Status: "queued"},
\t}
\tshipments := &mockShipmentRepo{}
\torders := &mockOrderRepo{order: &models.Order{ID: orderID, TenantID: tenantID}}

\tsvc := newShipmentSvc(carrier, shipments, orders, &models.Wallet{ID: uuid.New(), TenantID: tenantID})

\ts, err := svc.Dispatch(context.Background(), orderID, tenantID, terminalaf.BookRequest{})
\tif err != nil {
\t\tt.Fatalf("unexpected error: %v", err)
\t}
\tif s.CarrierRef == nil || *s.CarrierRef != "TA-001" {
\t\tt.Fatalf("expected CarrierRef=TA-001, got %v", s.CarrierRef)
\t}
\tif s.TrackingNumber == nil || *s.TrackingNumber != "TRK-001" {
\t\tt.Fatalf("expected TrackingNumber=TRK-001, got %v", s.TrackingNumber)
\t}
}

func TestDispatch_CarrierError(t *testing.T) {
\tcarrier := &mockCarrierClient{err: errors.New("carrier unavailable")}
\tsvc := newShipmentSvc(carrier, &mockShipmentRepo{}, &mockOrderRepo{}, &models.Wallet{ID: uuid.New()})

\t_, err := svc.Dispatch(context.Background(), uuid.New(), uuid.New(), terminalaf.BookRequest{})
\tif err == nil {
\t\tt.Fatal("expected error, got nil")
\t}
}

func TestHandleDelivered_ReleasesBalance(t *testing.T) {
\ttenantID := uuid.New()
\torderID := uuid.New()
\tshipmentID := uuid.New()

\torder := &models.Order{
\t\tID:          orderID,
\t\tTenantID:    tenantID,
\t\tTotalAmount: decimal.NewFromInt(5000),
\t\tShippingFee: decimal.NewFromInt(500),
\t}
\tshipment := &models.Shipment{ID: shipmentID, OrderID: orderID}
\twallet := &models.Wallet{
\t\tID:             uuid.New(),
\t\tTenantID:       tenantID,
\t\tPendingBalance: decimal.NewFromInt(5500),
\t}

\tshipments := &mockShipmentRepo{shipment: shipment}
\torders := &mockOrderRepo{order: order}
\tsvc := newShipmentSvc(&mockCarrierClient{}, shipments, orders, wallet)

\terr := svc.HandleDelivered(context.Background(), orderID)
\tif err != nil {
\t\tt.Fatalf("unexpected error: %v", err)
\t}
\tif shipments.statusSet != models.ShipmentStatusDelivered {
\t\tt.Fatalf("expected shipment status=delivered, got %s", shipments.statusSet)
\t}
}

func TestHandleDelivered_OrderNotFound(t *testing.T) {
\torders := &mockOrderRepo{err: errors.New("not found")}
\tsvc := newShipmentSvc(&mockCarrierClient{}, &mockShipmentRepo{}, orders, &models.Wallet{ID: uuid.New()})

\terr := svc.HandleDelivered(context.Background(), uuid.New())
\tif err == nil {
\t\tt.Fatal("expected error, got nil")
\t}
}
""")

write("internal/handler/webhook.go", """package handler

import (
\t"context"
\t"encoding/json"
\t"io"
\t"log/slog"
\t"net/http"

\t"github.com/google/uuid"
)

// chargeSuccessHandler is satisfied by *service.PaymentService.
type chargeSuccessHandler interface {
\tHandleChargeSuccess(ctx context.Context, reference string) error
}

// deliveredHandler is satisfied by *service.ShipmentService.
type deliveredHandler interface {
\tHandleDelivered(ctx context.Context, orderID uuid.UUID) error
}

type webhookVerifier interface {
\tVerifyWebhookSignature(payload []byte, signature string) bool
}

type WebhookHandler struct {
\tpaystackClient webhookVerifier
\tterminalClient webhookVerifier
\tpaymentSvc     chargeSuccessHandler
\tshipmentSvc    deliveredHandler
\tlog            *slog.Logger
}

func NewWebhookHandler(
\tpaystackClient webhookVerifier,
\tterminalClient webhookVerifier,
\tpaymentSvc chargeSuccessHandler,
\tshipmentSvc deliveredHandler,
\tlog *slog.Logger,
) *WebhookHandler {
\treturn &WebhookHandler{
\t\tpaystackClient: paystackClient,
\t\tterminalClient: terminalClient,
\t\tpaymentSvc:     paymentSvc,
\t\tshipmentSvc:    shipmentSvc,
\t\tlog:            log,
\t}
}

type incomingWebhookEvent struct {
\tEvent string          `json:"event"`
\tData  json.RawMessage `json:"data"`
}

// POST /webhooks/paystack
func (h *WebhookHandler) Paystack(w http.ResponseWriter, r *http.Request) {
\tbody, err := io.ReadAll(r.Body)
\tif err != nil {
\t\trespondErr(w, http.StatusBadRequest, "cannot read body")
\t\treturn
\t}

\tif !h.paystackClient.VerifyWebhookSignature(body, r.Header.Get("X-Paystack-Signature")) {
\t\trespondErr(w, http.StatusUnauthorized, "invalid signature")
\t\treturn
\t}

\tvar event incomingWebhookEvent
\tif err := json.Unmarshal(body, &event); err != nil {
\t\trespondErr(w, http.StatusBadRequest, "invalid event payload")
\t\treturn
\t}

\tif event.Event == "charge.success" {
\t\tvar data struct {
\t\t\tReference string `json:"reference"`
\t\t}
\t\tif err := json.Unmarshal(event.Data, &data); err == nil && data.Reference != "" {
\t\t\tif err := h.paymentSvc.HandleChargeSuccess(r.Context(), data.Reference); err != nil {
\t\t\t\th.log.Error("paystack webhook: charge.success", "error", err)
\t\t\t}
\t\t}
\t}

\tw.WriteHeader(http.StatusOK)
}

// POST /webhooks/terminalaf
func (h *WebhookHandler) TerminalAf(w http.ResponseWriter, r *http.Request) {
\tbody, err := io.ReadAll(r.Body)
\tif err != nil {
\t\trespondErr(w, http.StatusBadRequest, "cannot read body")
\t\treturn
\t}

\tif !h.terminalClient.VerifyWebhookSignature(body, r.Header.Get("X-Terminal-Africa-Signature")) {
\t\trespondErr(w, http.StatusUnauthorized, "invalid signature")
\t\treturn
\t}

\tvar event incomingWebhookEvent
\tif err := json.Unmarshal(body, &event); err != nil {
\t\trespondErr(w, http.StatusBadRequest, "invalid event payload")
\t\treturn
\t}

\tif event.Event == "shipment.delivered" {
\t\tvar data struct {
\t\t\tReference string `json:"metadata_reference"`
\t\t}
\t\tif err := json.Unmarshal(event.Data, &data); err == nil && data.Reference != "" {
\t\t\tif orderID, err := uuid.Parse(data.Reference); err == nil {
\t\t\t\tif err := h.shipmentSvc.HandleDelivered(r.Context(), orderID); err != nil {
\t\t\t\t\th.log.Error("terminalaf webhook: shipment.delivered", "error", err)
\t\t\t\t}
\t\t\t}
\t\t}
\t}

\tw.WriteHeader(http.StatusOK)
}
""")

write("internal/handler/webhook_test.go", """package handler_test

import (
\t"bytes"
\t"context"
\t"encoding/json"
\t"log/slog"
\t"net/http"
\t"net/http/httptest"
\t"testing"

\t"github.com/google/uuid"

\t"storefront/backend/internal/handler"
)

// ── stubs ─────────────────────────────────────────────────────────────────────

type stubWebhookVerifier struct{ valid bool }

func (s *stubWebhookVerifier) VerifyWebhookSignature(_ []byte, _ string) bool { return s.valid }

type stubPaymentWebhookSvc struct{ called bool }

func (s *stubPaymentWebhookSvc) HandleChargeSuccess(_ context.Context, _ string) error {
\ts.called = true
\treturn nil
}

type stubShipmentWebhookSvc struct{ called bool }

func (s *stubShipmentWebhookSvc) HandleDelivered(_ context.Context, _ uuid.UUID) error {
\ts.called = true
\treturn nil
}

func newWebhookHandler(validSig bool, payment *stubPaymentWebhookSvc, shipment *stubShipmentWebhookSvc) *handler.WebhookHandler {
\tv := &stubWebhookVerifier{valid: validSig}
\treturn handler.NewWebhookHandler(v, v, payment, shipment, slog.Default())
}

// ── paystack webhook ──────────────────────────────────────────────────────────

func TestPaystackWebhook_InvalidSignature(t *testing.T) {
\th := newWebhookHandler(false, &stubPaymentWebhookSvc{}, &stubShipmentWebhookSvc{})
\tbody, _ := json.Marshal(map[string]any{
\t\t"event": "charge.success",
\t\t"data":  map[string]any{"reference": uuid.New().String()},
\t})
\treq := httptest.NewRequest(http.MethodPost, "/webhooks/paystack", bytes.NewReader(body))
\trec := httptest.NewRecorder()
\th.Paystack(rec, req)
\tif rec.Code != http.StatusUnauthorized {
\t\tt.Fatalf("expected 401, got %d", rec.Code)
\t}
}

func TestPaystackWebhook_ChargeSuccess_Dispatches(t *testing.T) {
\tpaymentSvc := &stubPaymentWebhookSvc{}
\th := newWebhookHandler(true, paymentSvc, &stubShipmentWebhookSvc{})

\tinner, _ := json.Marshal(map[string]any{"reference": uuid.New().String()})
\tbody, _ := json.Marshal(map[string]any{"event": "charge.success", "data": json.RawMessage(inner)})
\treq := httptest.NewRequest(http.MethodPost, "/webhooks/paystack", bytes.NewReader(body))
\trec := httptest.NewRecorder()
\th.Paystack(rec, req)

\tif rec.Code != http.StatusOK {
\t\tt.Fatalf("expected 200, got %d", rec.Code)
\t}
\tif !paymentSvc.called {
\t\tt.Fatal("expected HandleChargeSuccess to be called")
\t}
}

func TestPaystackWebhook_UnknownEvent_NoDispatch(t *testing.T) {
\tpaymentSvc := &stubPaymentWebhookSvc{}
\th := newWebhookHandler(true, paymentSvc, &stubShipmentWebhookSvc{})

\tbody, _ := json.Marshal(map[string]any{"event": "transfer.success", "data": map[string]any{}})
\treq := httptest.NewRequest(http.MethodPost, "/webhooks/paystack", bytes.NewReader(body))
\trec := httptest.NewRecorder()
\th.Paystack(rec, req)

\tif rec.Code != http.StatusOK {
\t\tt.Fatalf("expected 200, got %d", rec.Code)
\t}
\tif paymentSvc.called {
\t\tt.Fatal("expected HandleChargeSuccess NOT to be called for unknown event")
\t}
}

// ── terminal africa webhook ───────────────────────────────────────────────────

func TestTerminalAfWebhook_InvalidSignature(t *testing.T) {
\th := newWebhookHandler(false, &stubPaymentWebhookSvc{}, &stubShipmentWebhookSvc{})
\tbody, _ := json.Marshal(map[string]any{"event": "shipment.delivered", "data": map[string]any{}})
\treq := httptest.NewRequest(http.MethodPost, "/webhooks/terminalaf", bytes.NewReader(body))
\trec := httptest.NewRecorder()
\th.TerminalAf(rec, req)
\tif rec.Code != http.StatusUnauthorized {
\t\tt.Fatalf("expected 401, got %d", rec.Code)
\t}
}

func TestTerminalAfWebhook_Delivered_Dispatches(t *testing.T) {
\tshipmentSvc := &stubShipmentWebhookSvc{}
\th := newWebhookHandler(true, &stubPaymentWebhookSvc{}, shipmentSvc)

\tinner, _ := json.Marshal(map[string]any{"metadata_reference": uuid.New().String()})
\tbody, _ := json.Marshal(map[string]any{"event": "shipment.delivered", "data": json.RawMessage(inner)})
\treq := httptest.NewRequest(http.MethodPost, "/webhooks/terminalaf", bytes.NewReader(body))
\trec := httptest.NewRecorder()
\th.TerminalAf(rec, req)

\tif rec.Code != http.StatusOK {
\t\tt.Fatalf("expected 200, got %d", rec.Code)
\t}
\tif !shipmentSvc.called {
\t\tt.Fatal("expected HandleDelivered to be called")
\t}
}
""")

print("Done.")

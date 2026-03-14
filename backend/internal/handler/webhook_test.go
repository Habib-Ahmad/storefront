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

	"storefront/backend/internal/handler"
)

// ── stubs ─────────────────────────────────────────────────────────────────────

type stubWebhookVerifier struct{ valid bool }

func (s *stubWebhookVerifier) VerifyWebhookSignature(_ []byte, _ string) bool { return s.valid }

type stubPaymentWebhookSvc struct {
	called       bool
	failedCalled bool
}

func (s *stubPaymentWebhookSvc) HandleChargeSuccess(_ context.Context, _ string) error {
	s.called = true
	return nil
}

func (s *stubPaymentWebhookSvc) HandleChargeFailed(_ context.Context, _ string) error {
	s.failedCalled = true
	return nil
}

type stubShipmentWebhookSvc struct {
	called       bool
	failedCalled bool
}

func (s *stubShipmentWebhookSvc) HandleDelivered(_ context.Context, _ uuid.UUID) error {
	s.called = true
	return nil
}

func (s *stubShipmentWebhookSvc) HandleShipmentFailed(_ context.Context, _ uuid.UUID) error {
	s.failedCalled = true
	return nil
}

func newWebhookHandler(validSig bool, payment *stubPaymentWebhookSvc, shipment *stubShipmentWebhookSvc) *handler.WebhookHandler {
	v := &stubWebhookVerifier{valid: validSig}
	return handler.NewWebhookHandler(v, v, payment, shipment, slog.Default())
}

// ── paystack webhook ──────────────────────────────────────────────────────────

func TestPaystackWebhook_InvalidSignature(t *testing.T) {
	h := newWebhookHandler(false, &stubPaymentWebhookSvc{}, &stubShipmentWebhookSvc{})
	body, _ := json.Marshal(map[string]any{
		"event": "charge.success",
		"data":  map[string]any{"reference": uuid.New().String()},
	})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/paystack", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Paystack(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPaystackWebhook_ChargeSuccess_Dispatches(t *testing.T) {
	paymentSvc := &stubPaymentWebhookSvc{}
	h := newWebhookHandler(true, paymentSvc, &stubShipmentWebhookSvc{})

	inner, _ := json.Marshal(map[string]any{"reference": uuid.New().String()})
	body, _ := json.Marshal(map[string]any{"event": "charge.success", "data": json.RawMessage(inner)})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/paystack", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Paystack(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !paymentSvc.called {
		t.Fatal("expected HandleChargeSuccess to be called")
	}
}

func TestPaystackWebhook_UnknownEvent_NoDispatch(t *testing.T) {
	paymentSvc := &stubPaymentWebhookSvc{}
	h := newWebhookHandler(true, paymentSvc, &stubShipmentWebhookSvc{})

	body, _ := json.Marshal(map[string]any{"event": "transfer.success", "data": map[string]any{}})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/paystack", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Paystack(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if paymentSvc.called {
		t.Fatal("expected HandleChargeSuccess NOT to be called for unknown event")
	}
}

// ── terminal africa webhook ───────────────────────────────────────────────────

func TestTerminalAfWebhook_InvalidSignature(t *testing.T) {
	h := newWebhookHandler(false, &stubPaymentWebhookSvc{}, &stubShipmentWebhookSvc{})
	body, _ := json.Marshal(map[string]any{"event": "shipment.delivered", "data": map[string]any{}})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/terminalaf", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.TerminalAf(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestTerminalAfWebhook_Delivered_Dispatches(t *testing.T) {
	shipmentSvc := &stubShipmentWebhookSvc{}
	h := newWebhookHandler(true, &stubPaymentWebhookSvc{}, shipmentSvc)

	inner, _ := json.Marshal(map[string]any{"metadata_reference": uuid.New().String()})
	body, _ := json.Marshal(map[string]any{"event": "shipment.delivered", "data": json.RawMessage(inner)})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/terminalaf", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.TerminalAf(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !shipmentSvc.called {
		t.Fatal("expected HandleDelivered to be called")
	}
}

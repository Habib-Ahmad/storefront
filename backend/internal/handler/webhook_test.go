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

	"github.com/google/uuid"

	"storefront/backend/internal/handler"
)

// ── stubs ─────────────────────────────────────────────────────────────────────

type stubWebhookVerifier struct{ valid bool }

func (s *stubWebhookVerifier) VerifyWebhookSignature(_ []byte, _ string) bool { return s.valid }

type stubPaymentWebhookSvc struct {
	called       bool
	failedCalled bool
	successErr   error
	failedErr    error
}

func (s *stubPaymentWebhookSvc) HandleChargeSuccess(_ context.Context, _ string) error {
	s.called = true
	return s.successErr
}

func (s *stubPaymentWebhookSvc) HandleChargeFailed(_ context.Context, _ string) error {
	s.failedCalled = true
	return s.failedErr
}

type stubShipmentWebhookSvc struct {
	called       bool
	failedCalled bool
	deliveredErr error
	failedErr    error
}

func (s *stubShipmentWebhookSvc) HandleDelivered(_ context.Context, _ uuid.UUID) error {
	s.called = true
	return s.deliveredErr
}

func (s *stubShipmentWebhookSvc) HandleShipmentFailed(_ context.Context, _ uuid.UUID) error {
	s.failedCalled = true
	return s.failedErr
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

func TestPaystackWebhook_ChargeSuccess_ProcessingFailureReturnsRetryableError(t *testing.T) {
	paymentSvc := &stubPaymentWebhookSvc{successErr: errors.New("db unavailable")}
	h := newWebhookHandler(true, paymentSvc, &stubShipmentWebhookSvc{})

	inner, _ := json.Marshal(map[string]any{"reference": uuid.New().String()})
	body, _ := json.Marshal(map[string]any{"event": "charge.success", "data": json.RawMessage(inner)})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/paystack", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Paystack(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestPaystackWebhook_InvalidChargePayload_ReturnsBadRequest(t *testing.T) {
	h := newWebhookHandler(true, &stubPaymentWebhookSvc{}, &stubShipmentWebhookSvc{})
	body, _ := json.Marshal(map[string]any{"event": "charge.success", "data": map[string]any{}})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/paystack", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Paystack(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
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

func TestTerminalAfWebhook_DeliveryProcessingFailureReturnsRetryableError(t *testing.T) {
	shipmentSvc := &stubShipmentWebhookSvc{deliveredErr: errors.New("wallet unavailable")}
	h := newWebhookHandler(true, &stubPaymentWebhookSvc{}, shipmentSvc)

	inner, _ := json.Marshal(map[string]any{"metadata_reference": uuid.New().String()})
	body, _ := json.Marshal(map[string]any{"event": "shipment.delivered", "data": json.RawMessage(inner)})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/terminalaf", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.TerminalAf(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestTerminalAfWebhook_InvalidReference_ReturnsBadRequest(t *testing.T) {
	h := newWebhookHandler(true, &stubPaymentWebhookSvc{}, &stubShipmentWebhookSvc{})

	inner, _ := json.Marshal(map[string]any{"metadata_reference": "not-a-uuid"})
	body, _ := json.Marshal(map[string]any{"event": "shipment.delivered", "data": json.RawMessage(inner)})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/terminalaf", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.TerminalAf(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

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

func newWebhookHandler(validSig bool, payment *stubPaymentWebhookSvc) *handler.WebhookHandler {
	v := &stubWebhookVerifier{valid: validSig}
	return handler.NewWebhookHandler(v, payment, slog.Default())
}

// ── paystack webhook ──────────────────────────────────────────────────────────

func TestPaystackWebhook_InvalidSignature(t *testing.T) {
	h := newWebhookHandler(false, &stubPaymentWebhookSvc{})
	body, _ := json.Marshal(map[string]any{
		"event": "charge.success",
		"data":  map[string]any{"reference": "ref-123"},
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
	h := newWebhookHandler(true, paymentSvc)

	inner, _ := json.Marshal(map[string]any{"reference": "ref-123"})
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
	h := newWebhookHandler(true, paymentSvc)

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
	h := newWebhookHandler(true, paymentSvc)

	inner, _ := json.Marshal(map[string]any{"reference": "ref-123"})
	body, _ := json.Marshal(map[string]any{"event": "charge.success", "data": json.RawMessage(inner)})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/paystack", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Paystack(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestPaystackWebhook_InvalidChargePayload_ReturnsBadRequest(t *testing.T) {
	h := newWebhookHandler(true, &stubPaymentWebhookSvc{})
	body, _ := json.Marshal(map[string]any{"event": "charge.success", "data": map[string]any{}})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/paystack", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Paystack(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

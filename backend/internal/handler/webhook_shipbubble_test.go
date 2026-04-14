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

type stubShipmentWebhookSvc struct {
	called     bool
	carrierRef string
	status     string
	payload    []byte
	err        error
}

func (s *stubShipmentWebhookSvc) HandleStatusUpdate(_ context.Context, carrierRef, status string, payload []byte) error {
	s.called = true
	s.carrierRef = carrierRef
	s.status = status
	s.payload = append([]byte(nil), payload...)
	return s.err
}

func TestShipbubbleWebhook_InvalidSignature(t *testing.T) {
	h := handler.NewWebhookHandler(&stubWebhookVerifier{valid: true}, &stubPaymentWebhookSvc{}, slog.Default())
	h.SetShipmentService(&stubWebhookVerifier{valid: false}, &stubShipmentWebhookSvc{})
	body, _ := json.Marshal(map[string]any{"event": "shipment.label.created", "order_id": "SB-123", "status": "pending"})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/shipbubble", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Shipbubble(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestShipbubbleWebhook_ShipmentEvent_Dispatches(t *testing.T) {
	shipmentSvc := &stubShipmentWebhookSvc{}
	h := handler.NewWebhookHandler(&stubWebhookVerifier{valid: true}, &stubPaymentWebhookSvc{}, slog.Default())
	h.SetShipmentService(&stubWebhookVerifier{valid: true}, shipmentSvc)
	body, _ := json.Marshal(map[string]any{"event": "shipment.label.created", "order_id": "SB-123", "status": "pending"})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/shipbubble", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Shipbubble(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !shipmentSvc.called {
		t.Fatal("expected shipment webhook handler to be called")
	}
	if shipmentSvc.carrierRef != "SB-123" || shipmentSvc.status != "pending" {
		t.Fatalf("unexpected shipment webhook args: %+v", shipmentSvc)
	}
}

func TestShipbubbleWebhook_ProcessingFailureReturnsRetryableError(t *testing.T) {
	shipmentSvc := &stubShipmentWebhookSvc{err: errors.New("db unavailable")}
	h := handler.NewWebhookHandler(&stubWebhookVerifier{valid: true}, &stubPaymentWebhookSvc{}, slog.Default())
	h.SetShipmentService(&stubWebhookVerifier{valid: true}, shipmentSvc)
	body, _ := json.Marshal(map[string]any{"event": "shipment.label.created", "order_id": "SB-123", "status": "pending"})
	req := httptest.NewRequest(http.MethodPost, "/webhooks/shipbubble", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	h.Shipbubble(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

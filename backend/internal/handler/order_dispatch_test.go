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
	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

type stubShipmentDispatcher struct {
	shipment *models.Shipment
	options  []models.DispatchShipmentOption
	err      error
	called   bool
	quoted   bool
	orderID  uuid.UUID
	tenantID uuid.UUID
	req      service.DispatchShipmentRequest
}

func (s *stubShipmentDispatcher) QuoteDispatchOptions(_ context.Context, orderID, tenantID uuid.UUID) ([]models.DispatchShipmentOption, error) {
	s.quoted = true
	s.orderID = orderID
	s.tenantID = tenantID
	if s.err != nil {
		return nil, s.err
	}
	return s.options, nil
}

func (s *stubShipmentDispatcher) Dispatch(_ context.Context, orderID, tenantID uuid.UUID, req service.DispatchShipmentRequest) (*models.Shipment, error) {
	s.called = true
	s.orderID = orderID
	s.tenantID = tenantID
	s.req = req
	if s.err != nil {
		return nil, s.err
	}
	if s.shipment != nil {
		return s.shipment, nil
	}
	return &models.Shipment{ID: uuid.New()}, nil
}

func TestDispatchOrder_CallsShipmentService(t *testing.T) {
	variantID := uuid.New()
	orderID := uuid.New()
	tenantID := uuid.New()
	dispatcher := &stubShipmentDispatcher{shipment: &models.Shipment{ID: uuid.New(), OrderID: orderID, TenantID: tenantID}}
	h := handler.NewOrderHandler(service.NewOrderService(&stubOrderRepo{}, &stubProductRepoForOrder{variant: &models.ProductVariant{ID: variantID}}), &stubPaymentInitiator{}, "", slog.Default())
	h.SetShipmentService(dispatcher)

	body, _ := json.Marshal(map[string]any{
		"courier_id":   "123",
		"service_code": "bike",
		"service_type": "dropoff",
	})
	req := httptest.NewRequest(http.MethodPost, "/orders/"+orderID.String()+"/dispatch", bytes.NewReader(body))
	req = withURLParam(req, "id", orderID.String())
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: tenantID, ActiveModules: models.ActiveModules{Logistics: true}}))
	rec := httptest.NewRecorder()

	h.Dispatch(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if !dispatcher.called {
		t.Fatal("expected shipment dispatch service to be called")
	}
	if dispatcher.orderID != orderID {
		t.Fatalf("order_id: want %s, got %s", orderID, dispatcher.orderID)
	}
	if dispatcher.tenantID != tenantID {
		t.Fatalf("tenant_id: want %s, got %s", tenantID, dispatcher.tenantID)
	}
	if dispatcher.req.CourierID != "123" || dispatcher.req.ServiceCode != "bike" || dispatcher.req.ServiceType != "dropoff" {
		t.Fatalf("unexpected dispatch request: %+v", dispatcher.req)
	}
}

func TestDispatchOrder_RequiresLogisticsModule(t *testing.T) {
	h := handler.NewOrderHandler(service.NewOrderService(&stubOrderRepo{}, &stubProductRepoForOrder{}), &stubPaymentInitiator{}, "", slog.Default())
	h.SetShipmentService(&stubShipmentDispatcher{})

	req := httptest.NewRequest(http.MethodPost, "/orders/"+uuid.New().String()+"/dispatch", bytes.NewReader([]byte(`{"courier_id":"123","service_code":"bike"}`)))
	req = withURLParam(req, "id", uuid.New().String())
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: uuid.New(), ActiveModules: models.ActiveModules{Payments: true}}))
	rec := httptest.NewRecorder()

	h.Dispatch(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDispatchOptions_ReturnsCourierOptions(t *testing.T) {
	orderID := uuid.New()
	tenantID := uuid.New()
	dispatcher := &stubShipmentDispatcher{options: []models.DispatchShipmentOption{{
		ID:          "123:bike:dropoff",
		CourierID:   "123",
		CourierName: "Kwik",
		ServiceCode: "bike",
		ServiceType: "dropoff",
		Amount:      "1500",
		Currency:    "NGN",
	}}}
	h := handler.NewOrderHandler(service.NewOrderService(&stubOrderRepo{}, &stubProductRepoForOrder{}), &stubPaymentInitiator{}, "", slog.Default())
	h.SetShipmentService(dispatcher)

	req := httptest.NewRequest(http.MethodGet, "/orders/"+orderID.String()+"/dispatch-options", nil)
	req = withURLParam(req, "id", orderID.String())
	req = req.WithContext(injectTenant(req.Context(), &models.Tenant{ID: tenantID, ActiveModules: models.ActiveModules{Logistics: true}}))
	rec := httptest.NewRecorder()

	h.DispatchOptions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !dispatcher.quoted {
		t.Fatal("expected dispatch options service to be called")
	}
	if dispatcher.orderID != orderID || dispatcher.tenantID != tenantID {
		t.Fatalf("unexpected dispatch-options args: order=%s tenant=%s", dispatcher.orderID, dispatcher.tenantID)
	}
}

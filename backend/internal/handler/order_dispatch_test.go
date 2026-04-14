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
	err      error
	called   bool
	orderID  uuid.UUID
	tenantID uuid.UUID
	req      service.DispatchShipmentRequest
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

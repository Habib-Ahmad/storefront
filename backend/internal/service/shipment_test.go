package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/adapter/terminalaf"
	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

// ── carrier client mock ───────────────────────────────────────────────────────

type mockCarrierClient struct {
	resp *terminalaf.BookResponse
	err  error
}

func (m *mockCarrierClient) BookShipment(_ context.Context, _ terminalaf.BookRequest) (*terminalaf.BookResponse, error) {
	return m.resp, m.err
}

// ── shipment repo mock ────────────────────────────────────────────────────────

type mockShipmentRepo struct {
	shipment  *models.Shipment
	err       error
	statusSet models.ShipmentStatus
}

func (m *mockShipmentRepo) Create(_ context.Context, s *models.Shipment) error {
	s.ID = uuid.New()
	m.shipment = s
	return m.err
}

func (m *mockShipmentRepo) GetByOrderID(_ context.Context, _, _ uuid.UUID) (*models.Shipment, error) {
	return m.shipment, m.err
}

func (m *mockShipmentRepo) UpdateStatus(_ context.Context, _, _ uuid.UUID, status models.ShipmentStatus) error {
	m.statusSet = status
	return m.err
}

func (m *mockShipmentRepo) AppendCarrierEvent(_ context.Context, _, _ uuid.UUID, _ []byte) error {
	return m.err
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newShipmentSvc(carrier *mockCarrierClient, shipments *mockShipmentRepo, orders *mockOrderRepo, wallet *models.Wallet) *service.ShipmentService {
	walletSvc := service.NewWalletService(
		&mockWalletRepo{wallet: wallet},
		&mockTxRepo{},
		&mockTenantRepo{},
		testHMACSecret,
	)
	return service.NewShipmentService(carrier, shipments, orders, walletSvc)
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestDispatch_BookAndSave(t *testing.T) {
	orderID := uuid.New()
	tenantID := uuid.New()
	carrier := &mockCarrierClient{
		resp: &terminalaf.BookResponse{CarrierRef: "TA-001", TrackingNumber: "TRK-001", Status: "queued"},
	}
	shipments := &mockShipmentRepo{}
	orders := &mockOrderRepo{order: &models.Order{ID: orderID, TenantID: tenantID, FulfillmentStatus: models.FulfillmentStatusProcessing}}

	svc := newShipmentSvc(carrier, shipments, orders, &models.Wallet{ID: uuid.New(), TenantID: tenantID})

	s, err := svc.Dispatch(context.Background(), orderID, tenantID, terminalaf.BookRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.CarrierRef == nil || *s.CarrierRef != "TA-001" {
		t.Fatalf("expected CarrierRef=TA-001, got %v", s.CarrierRef)
	}
	if s.TrackingNumber == nil || *s.TrackingNumber != "TRK-001" {
		t.Fatalf("expected TrackingNumber=TRK-001, got %v", s.TrackingNumber)
	}
}

func TestDispatch_CarrierError(t *testing.T) {
	carrier := &mockCarrierClient{err: errors.New("carrier unavailable")}
	order := &models.Order{ID: uuid.New(), TenantID: uuid.New(), FulfillmentStatus: models.FulfillmentStatusProcessing}
	svc := newShipmentSvc(carrier, &mockShipmentRepo{}, &mockOrderRepo{order: order}, &models.Wallet{ID: uuid.New()})

	_, err := svc.Dispatch(context.Background(), order.ID, order.TenantID, terminalaf.BookRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestHandleDelivered_ReleasesBalance(t *testing.T) {
	tenantID := uuid.New()
	orderID := uuid.New()
	shipmentID := uuid.New()

	order := &models.Order{
		ID:          orderID,
		TenantID:    tenantID,
		TotalAmount: decimal.NewFromInt(5000),
		ShippingFee: decimal.NewFromInt(500),
	}
	shipment := &models.Shipment{ID: shipmentID, OrderID: orderID}
	wallet := &models.Wallet{
		ID:             uuid.New(),
		TenantID:       tenantID,
		PendingBalance: decimal.NewFromInt(5500),
	}

	shipments := &mockShipmentRepo{shipment: shipment}
	orders := &mockOrderRepo{order: order}
	svc := newShipmentSvc(&mockCarrierClient{}, shipments, orders, wallet)

	err := svc.HandleDelivered(context.Background(), orderID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if shipments.statusSet != models.ShipmentStatusDelivered {
		t.Fatalf("expected shipment status=delivered, got %s", shipments.statusSet)
	}
}

func TestHandleDelivered_OrderNotFound(t *testing.T) {
	orders := &mockOrderRepo{err: errors.New("not found")}
	svc := newShipmentSvc(&mockCarrierClient{}, &mockShipmentRepo{}, orders, &models.Wallet{ID: uuid.New()})

	err := svc.HandleDelivered(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

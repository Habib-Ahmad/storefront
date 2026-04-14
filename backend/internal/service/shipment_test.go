package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"storefront/backend/internal/adapter/shipbubble"
	"storefront/backend/internal/models"
	"storefront/backend/internal/service"
)

func shipmentStringPtr(value string) *string { return &value }

type mockShipmentRepo struct {
	shipment      *models.Shipment
	err           error
	statusSet     models.ShipmentStatus
	upserted      *models.Shipment
	appendedEvent []byte
	carrierRef    string
}

func (m *mockShipmentRepo) Create(_ context.Context, s *models.Shipment) error {
	s.ID = uuid.New()
	m.shipment = s
	return m.err
}

func (m *mockShipmentRepo) UpsertDispatch(_ context.Context, s *models.Shipment) error {
	s.ID = uuid.New()
	m.upserted = s
	m.shipment = s
	return m.err
}

func (m *mockShipmentRepo) GetByOrderID(_ context.Context, _, _ uuid.UUID) (*models.Shipment, error) {
	return m.shipment, m.err
}

func (m *mockShipmentRepo) GetByCarrierRef(_ context.Context, carrierRef string) (*models.Shipment, error) {
	m.carrierRef = carrierRef
	return m.shipment, m.err
}

func (m *mockShipmentRepo) UpdateStatus(_ context.Context, _, _ uuid.UUID, status models.ShipmentStatus) error {
	m.statusSet = status
	if m.shipment != nil {
		m.shipment.Status = status
	}
	return m.err
}

func (m *mockShipmentRepo) AppendCarrierEvent(_ context.Context, _, _ uuid.UUID, event []byte) error {
	m.appendedEvent = append([]byte(nil), event...)
	return m.err
}

type stubShipmentProvider struct {
	validatedAddress *shipbubble.ValidatedAddress
	categories       []shipbubble.PackageCategory
	boxes            []shipbubble.PackageBox
	rateResponse     *shipbubble.RateResponse
	shipmentRecord   *shipbubble.ShipmentRecord
	createErr        error
	lastCreate       shipbubble.CreateShipmentRequest
	lastRate         shipbubble.RateRequest
}

func (s *stubShipmentProvider) ValidateAddress(_ context.Context, _ shipbubble.ValidateAddressRequest) (*shipbubble.ValidatedAddress, error) {
	if s.validatedAddress != nil {
		copy := *s.validatedAddress
		return &copy, nil
	}
	return &shipbubble.ValidatedAddress{AddressCode: 1001}, nil
}

func (s *stubShipmentProvider) GetPackageCategories(_ context.Context) ([]shipbubble.PackageCategory, error) {
	return s.categories, nil
}

func (s *stubShipmentProvider) GetPackageBoxes(_ context.Context) ([]shipbubble.PackageBox, error) {
	return s.boxes, nil
}

func (s *stubShipmentProvider) FetchRates(_ context.Context, req shipbubble.RateRequest) (*shipbubble.RateResponse, error) {
	s.lastRate = req
	return s.rateResponse, nil
}

func (s *stubShipmentProvider) CreateShipment(_ context.Context, req shipbubble.CreateShipmentRequest) (*shipbubble.ShipmentRecord, error) {
	s.lastCreate = req
	if s.createErr != nil {
		return nil, s.createErr
	}
	return s.shipmentRecord, nil
}

func TestDispatch_BooksAndStoresShipbubbleShipment(t *testing.T) {
	tenantID := uuid.New()
	orderID := uuid.New()
	variantID := uuid.New()
	productID := uuid.New()
	tenantPhone := "+2348012345678"
	tenantAddress := "12 Allen Avenue, Ikeja"

	orders := &mockOrderRepo{order: &models.Order{
		ID:                orderID,
		TenantID:          tenantID,
		IsDelivery:        true,
		PaymentMethod:     models.PaymentMethodOnline,
		PaymentStatus:     models.PaymentStatusPaid,
		FulfillmentStatus: models.FulfillmentStatusProcessing,
		CustomerName:      shipmentStringPtr("Ada"),
		CustomerPhone:     shipmentStringPtr("08012345678"),
		CustomerEmail:     shipmentStringPtr("ada@example.com"),
		ShippingAddress:   shipmentStringPtr("23 Broad Street, Lagos"),
		TotalAmount:       decimal.NewFromInt(5000),
		ShippingFee:       decimal.NewFromInt(1500),
		CreatedAt:         time.Now(),
	}}
	orders.items = []models.OrderItem{{
		VariantID:    variantID,
		Quantity:     2,
		PriceAtSale:  decimal.NewFromInt(2500),
		ProductName:  shipmentStringPtr("Ankara Set"),
		VariantLabel: shipmentStringPtr("default"),
	}}
	products := &mockProductRepo{
		product: &models.Product{
			ID:          productID,
			TenantID:    tenantID,
			Name:        "Ankara Set",
			Description: shipmentStringPtr("A bright two-piece set"),
			Category:    shipmentStringPtr("Fashion"),
			IsAvailable: true,
		},
		variant: &models.ProductVariant{ID: variantID, ProductID: productID},
	}
	tenants := &mockTenantRepo{tenant: &models.Tenant{
		ID:           tenantID,
		Name:         "Funke Fabrics",
		Slug:         "funke-fabrics",
		ContactPhone: &tenantPhone,
		Address:      &tenantAddress,
	}}
	provider := &stubShipmentProvider{
		validatedAddress: &shipbubble.ValidatedAddress{AddressCode: 1001},
		categories:       []shipbubble.PackageCategory{{ID: 2, Name: "Fashion wears"}},
		boxes:            []shipbubble.PackageBox{{Name: "medium box", Length: decimal.NewFromInt(16), Width: decimal.NewFromInt(12), Height: decimal.NewFromInt(10), MaxWeight: decimal.RequireFromString("2.00")}},
		rateResponse: &shipbubble.RateResponse{
			RequestToken: "req-token",
			Options: []shipbubble.RateOption{{
				CourierID:   "123",
				ServiceCode: "bike",
				ServiceType: "dropoff",
			}},
		},
		shipmentRecord: &shipbubble.ShipmentRecord{
			OrderID: "SB-123",
			Status:  "pending",
			Courier: shipbubble.ShipmentCourier{TrackingCode: "TRK-123"},
			Raw:     []byte(`{"order_id":"SB-123"}`),
		},
	}
	shipments := &mockShipmentRepo{}

	svc := service.NewShipmentService(provider, shipments, orders, products, tenants, nil)
	shipment, err := svc.Dispatch(context.Background(), orderID, tenantID, service.DispatchShipmentRequest{
		CourierID:   "123",
		ServiceCode: "bike",
		ServiceType: "dropoff",
	})
	if err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}
	if shipment.CarrierRef == nil || *shipment.CarrierRef != "SB-123" {
		t.Fatalf("carrier_ref: want SB-123, got %v", shipment.CarrierRef)
	}
	if provider.lastCreate.RequestToken != "req-token" {
		t.Fatalf("request_token: want req-token, got %s", provider.lastCreate.RequestToken)
	}
	if orders.fulfillmentStatus != models.FulfillmentStatusShipped {
		t.Fatalf("fulfillment_status: want shipped, got %s", orders.fulfillmentStatus)
	}
	if shipments.upserted == nil {
		t.Fatal("expected shipment booking to be persisted")
	}
	if shipments.upserted.Status != models.ShipmentStatusQueued {
		t.Fatalf("shipment status: want queued, got %s", shipments.upserted.Status)
	}
}

func TestDispatch_RequiresPaidDeliveryOrder(t *testing.T) {
	order := &models.Order{
		ID:                uuid.New(),
		TenantID:          uuid.New(),
		IsDelivery:        true,
		PaymentMethod:     models.PaymentMethodOnline,
		PaymentStatus:     models.PaymentStatusPending,
		FulfillmentStatus: models.FulfillmentStatusProcessing,
	}
	svc := service.NewShipmentService(&stubShipmentProvider{}, &mockShipmentRepo{}, &mockOrderRepo{order: order}, &mockProductRepo{}, &mockTenantRepo{}, nil)

	_, err := svc.Dispatch(context.Background(), order.ID, order.TenantID, service.DispatchShipmentRequest{CourierID: "123", ServiceCode: "bike"})
	if err == nil {
		t.Fatal("expected dispatch to reject unpaid delivery orders")
	}
}

func TestHandleStatusUpdate_Delivered_ReleasesPendingBalance(t *testing.T) {
	tenantID := uuid.New()
	orderID := uuid.New()
	shipmentID := uuid.New()
	shipments := &mockShipmentRepo{shipment: &models.Shipment{ID: shipmentID, OrderID: orderID, TenantID: tenantID, CarrierRef: shipmentStringPtr("SB-123")}}
	orders := &mockOrderRepo{order: &models.Order{
		ID:                orderID,
		TenantID:          tenantID,
		IsDelivery:        true,
		PaymentMethod:     models.PaymentMethodOnline,
		PaymentStatus:     models.PaymentStatusPaid,
		FulfillmentStatus: models.FulfillmentStatusShipped,
		TotalAmount:       decimal.NewFromInt(5000),
		ShippingFee:       decimal.NewFromInt(500),
	}}
	walletRepo := &mockWalletRepo{wallet: &models.Wallet{ID: uuid.New(), TenantID: tenantID, PendingBalance: decimal.NewFromInt(5500)}}
	walletSvc := service.NewWalletService(walletRepo, &mockTxRepo{}, &mockTenantRepo{}, testHMACSecret)
	svc := service.NewShipmentService(nil, shipments, orders, &mockProductRepo{}, &mockTenantRepo{}, walletSvc)

	err := svc.HandleStatusUpdate(context.Background(), "SB-123", "completed", []byte(`{"status":"completed"}`))
	if err != nil {
		t.Fatalf("HandleStatusUpdate returned error: %v", err)
	}
	if shipments.statusSet != models.ShipmentStatusDelivered {
		t.Fatalf("shipment status: want delivered, got %s", shipments.statusSet)
	}
	if orders.fulfillmentStatus != models.FulfillmentStatusDelivered {
		t.Fatalf("fulfillment_status: want delivered, got %s", orders.fulfillmentStatus)
	}
	if walletRepo.updated == nil || !walletRepo.updated.AvailableBalance.Equal(decimal.NewFromInt(5500)) {
		t.Fatalf("available_balance: want 5500, got %v", walletRepo.updated)
	}
	if len(shipments.appendedEvent) == 0 {
		t.Fatal("expected webhook payload to be appended to shipment history")
	}
}

func TestHandleStatusUpdate_Cancelled_ResetsOrderForRedispatch(t *testing.T) {
	tenantID := uuid.New()
	orderID := uuid.New()
	shipmentID := uuid.New()
	shipments := &mockShipmentRepo{shipment: &models.Shipment{ID: shipmentID, OrderID: orderID, TenantID: tenantID, CarrierRef: shipmentStringPtr("SB-124")}}
	orders := &mockOrderRepo{order: &models.Order{
		ID:                orderID,
		TenantID:          tenantID,
		IsDelivery:        true,
		PaymentMethod:     models.PaymentMethodOnline,
		PaymentStatus:     models.PaymentStatusPaid,
		FulfillmentStatus: models.FulfillmentStatusShipped,
	}}
	svc := service.NewShipmentService(nil, shipments, orders, &mockProductRepo{}, &mockTenantRepo{}, nil)

	err := svc.HandleStatusUpdate(context.Background(), "SB-124", "cancelled", []byte(`{"status":"cancelled"}`))
	if err != nil {
		t.Fatalf("HandleStatusUpdate returned error: %v", err)
	}
	if shipments.statusSet != models.ShipmentStatusFailed {
		t.Fatalf("shipment status: want failed, got %s", shipments.statusSet)
	}
	if orders.fulfillmentStatus != models.FulfillmentStatusProcessing {
		t.Fatalf("fulfillment_status: want processing, got %s", orders.fulfillmentStatus)
	}
}

func TestHandleStatusUpdate_UnknownShipmentReturnsError(t *testing.T) {
	svc := service.NewShipmentService(nil, &mockShipmentRepo{err: errors.New("missing")}, &mockOrderRepo{}, &mockProductRepo{}, &mockTenantRepo{}, nil)
	if err := svc.HandleStatusUpdate(context.Background(), "SB-missing", "completed", nil); err == nil {
		t.Fatal("expected missing shipment to return an error")
	}
}
